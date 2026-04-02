import base64
import json
import logging
import requests
from io import BytesIO
from google import genai
from groq import Groq
from PIL import Image
from mindee import (
    ClientV2,
    InferenceParameters,
    InferenceResponse,
    BytesInput,
)
from utils.categoriesList import categoriesList
from models.uploadModels import UploadReceiptResponse, ReceiptItem

logger = logging.getLogger("genai.cv")


class ProcessReceipts:
    def __init__(self, ocrApiKey, modelId, geminiApiKey, groqApiKey, geminiModel, groqModel):
        # api keys
        self.ocrApiKey = ocrApiKey
        self.geminiApiKey = geminiApiKey
        self.groqApiKey = groqApiKey
        
        # model versions
        self.ocrModelId = modelId
        self.geminiModel = geminiModel
        self.groqModel = groqModel

        # initialising clients
        self.ocrClient = ClientV2(self.ocrApiKey)
        self.geminiClient = genai.Client(api_key=self.geminiApiKey)
        self.groqClient = Groq(api_key=self.groqApiKey)

    def convertImageToData(self, image: str, currency: str) -> dict:
        # Decode the base64 image
        imageBytes = base64.b64decode(image)
        pilImage = Image.open(BytesIO(imageBytes))
        extension = (pilImage.format or "jpg").lower()

        logger.info("Image decoded | Format=%s SizeBytes=%d", extension, len(imageBytes))

        # Set up Mindee OCR params
        params = InferenceParameters(
            model_id=self.ocrModelId,
            rag=None,
            raw_text=None,
            polygon=None,
            confidence=True
        )

        inputSource = BytesInput(imageBytes, filename=f"receipt.{extension}")

        # Call Mindee OCR
        logger.info("Calling Mindee OCR")
        try:
            result = self.ocrClient.enqueue_and_get_result(
                InferenceResponse,
                inputSource,
                params
            )
        except Exception as e:
            logger.error("Mindee OCR call failed | Error=%s", str(e))
            raise RuntimeError(f"Mindee OCR call failed: {e}")

        # Extract fields
        fields = result.inference.result.fields
        merchantName = fields.get("supplier_name").value
        date = fields.get("date").value
        totalAmount = fields.get("total_amount").value
        currencyCode = fields["locale"].fields["currency"].value

        logger.info("OCR extraction complete | Merchant=%s Date=%s Total=%s Currency=%s",
                     merchantName, date, totalAmount, currencyCode)

        # Calculate confidence score
        confidenceMap = {"Certain": 1.0, "High": 0.85, "Medium": 0.6, "Low": 0.3}
        keyFields = ["supplier_name", "date", "total_amount"]
        confidenceValues = []
        for key in keyFields:
            field = fields.get(key)
            if field and field.confidence:
                confidenceValues.append(confidenceMap.get(str(field.confidence), 0.0))
        confidenceScore = sum(confidenceValues) / len(confidenceValues) if confidenceValues else 0.0

        logger.info("Confidence score calculated | Score=%.2f", confidenceScore)

        # Extract line items
        items = fields.get("line_items").items
        itemsList: list = []

        for item in items:
            itemFields = item.fields
            itemName = itemFields["description"].value
            itemQuantity = itemFields["quantity"].value
            itemPrice = itemFields["total_price"].value

            itemObject: dict = {
                "name": itemName,
                "price": itemPrice,
                "quantity": itemQuantity if itemQuantity is not None else 1
            }
            itemsList.append(itemObject)

        logger.info("Line items extracted | Count=%d", len(itemsList))

        # Categorize items via LLM
        itemsList: ReceiptItem = self.addCategoriesToList(itemsList)

        responseData: UploadReceiptResponse = {
            "merchant": merchantName,
            "date": date,
            "totalAmount": totalAmount,
            "currency": currencyCode,
            "items": itemsList,
            "confidenceScore": confidenceScore
        }
        return responseData

    def addCategoriesToList(self, itemsList: list) -> list:
        categories = ", ".join(categoriesList)
        itemNames = [item["name"] for item in itemsList]

        prompt = (
            f"You are a categorization assistant. Given a list of item names, assign each item a category "
            f"from ONLY this list: [{categories}].\n\n"
            f"Items: {json.dumps(itemNames)}\n\n"
            f"Respond with a JSON array of objects with 'name' and 'category' fields. "
            f"No explanation, just the JSON array."
        )

        if not hasattr(self, "_llm_toggle"):
            self._llm_toggle = True

        if self._llm_toggle:
            primary, fallback = self._call_gemini, self._call_groq
            primaryName, fallbackName = "Gemini", "Groq"
        else:
            primary, fallback = self._call_groq, self._call_gemini
            primaryName, fallbackName = "Groq", "Gemini"
        self._llm_toggle = not self._llm_toggle

        logger.info("Categorizing %d items | Primary=%s Fallback=%s", len(itemNames), primaryName, fallbackName)

        result = primary(prompt)
        if not result:
            logger.warn("Primary LLM (%s) failed, trying fallback (%s)", primaryName, fallbackName)
            result = fallback(prompt)

        if result:
            try:
                parsed = json.loads(result)
                category_map = {entry["name"]: entry["category"] for entry in parsed}
                for item in itemsList:
                    item["category"] = category_map.get(item["name"], "Miscellaneous")
                logger.info("LLM categorization successful | CategoriesAssigned=%d", len(category_map))
            except (json.JSONDecodeError, KeyError) as e:
                logger.error("Failed to parse LLM response | Error=%s", str(e))
                for item in itemsList:
                    item["category"] = "Miscellaneous"
        else:
            logger.error("Both LLM providers failed, defaulting all items to Miscellaneous")
            for item in itemsList:
                item["category"] = "Miscellaneous"

        return itemsList

    def _call_gemini(self, prompt: str) -> str | None:
        try:
            logger.info("Calling Gemini API")
            response = self.geminiClient.models.generate_content(
                model=self.geminiModel,
                contents=prompt
            )
            logger.info("Gemini API responded successfully")
            return response.text.strip()
        except Exception as e:
            logger.error("Gemini API call failed | Error=%s", str(e))
            return None

    def _call_groq(self, prompt: str) -> str | None:
        try:
            logger.info("Calling Groq API")
            response = self.groqClient.chat.completions.create(
                model=self.groqModel,
                messages=[{"role": "user", "content": prompt}],
                max_tokens=1024
            )
            logger.info("Groq API responded successfully")
            return response.choices[0].message.content.strip()
        except Exception as e:
            logger.error("Groq API call failed | Error=%s", str(e))
            return None
