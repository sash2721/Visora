import base64
import json
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
from configs.categoriesList import categoriesList
from models.uploadModels import UploadReceiptResponse, ReceiptItem

class ProcessReceipts:
    def __init__(self, ocrApiKey, modelId, geminiApiKey, groqApiKey):
        self.ocrApiKey = ocrApiKey
        self.geminiApiKey = geminiApiKey
        self.groqApiKey = groqApiKey
        self.ocrClient = ClientV2(self.ocrApiKey)
        self.geminiClient = genai.Client(api_key=self.geminiApiKey)
        self.groqClient = Groq(api_key=self.groqApiKey)
        self.modelId = modelId

    def convertImageToData(self, image: str, currency: str) -> dict:
        # Creating the image from the bytes sent by the backend
        imageBytes = base64.b64decode(image)
        pilImage = Image.open(BytesIO(imageBytes))

        # Getting the extension of the file sent by the backend
        extension = (pilImage.format or "jpg").lower()

        # Implementing Mindee OCR to extract receipt data
        # defining the right params for the model
        params = InferenceParameters(
            model_id=modelId,
            rag=None,
            raw_text=None,
            polygon=None,
            confidence=True
        )
        
        # defining the image path as input source
        inputSource = BytesInput(imageBytes, filename=f"receipt.{extension}")

        # calling the API for the given image
        result = self.ocrClient.enqueue_and_get_result(
            InferenceResponse,
            inputScore,
            params
        )

        # extracting the fields
        fields = result.inference.result.fields
        merchantName = fields.get("supplier_name").value
        date = fields.get("date").value
        totalAmount = fields.get("total_amount").value
        currencyCode = fields["locale"].fields["currency"].value

        # calculating overall confidence score from key fields
        confidenceMap = {"Certain": 1.0, "High": 0.85, "Medium": 0.6, "Low": 0.3}
        keyFields = ["supplier_name", "date", "total_amount"]
        confidenceValues = []
        for key in keyFields:
            field = fields.get(key)
            if field and field.confidence:
                confidenceValues.append(confidenceMap.get(str(field.confidence), 0.0))
        confidenceScore = sum(confidenceValues) / len(confidenceValues) if confidenceValues else 0.0

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
                "quantity": itemQuantity
            }
            itemsList.append(itemObject)
        
        # sending the items list to LLM to add category to each item in the list
        itemsList: ReceiptItem = self.addCategoriesToList(itemsList)

        # Send the response back to the backend
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

        # Alternate primary based on a simple toggle
        if not hasattr(self, "_llm_toggle"):
            self._llm_toggle = True

        if self._llm_toggle:
            primary, fallback = self._call_gemini, self._call_groq
        else:
            primary, fallback = self._call_groq, self._call_gemini
        self._llm_toggle = not self._llm_toggle

        result = primary(prompt)
        if not result:
            result = fallback(prompt)

        if result:
            try:
                parsed = json.loads(result)
                category_map = {entry["name"]: entry["category"] for entry in parsed}
                for item in itemsList:
                    item["category"] = category_map.get(item["name"], "Miscellaneous")
            except (json.JSONDecodeError, KeyError):
                for item in itemsList:
                    item["category"] = "Miscellaneous"
        else:
            for item in itemsList:
                item["category"] = "Miscellaneous"

        return itemsList

    def _call_gemini(self, prompt: str) -> str | None:
        try:
            response = self.geminiClient.models.generate_content(
                model="gemini-2.0-flash",
                contents=prompt
            )
            return response.text.strip()
        except Exception:
            return None

    def _call_groq(self, prompt: str) -> str | None:
        try:
            response = self.groqClient.chat.completions.create(
                model="llama-3.3-70b-versatile",
                messages=[{"role": "user", "content": prompt}],
                max_tokens=1024
            )
            return response.choices[0].message.content.strip()
        except Exception:
            return None