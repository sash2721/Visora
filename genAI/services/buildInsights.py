import json
import logging

from google import genai
from groq import Groq

logger = logging.getLogger("genai")

_next_provider = "groq"

class BuildInsights:
    def __init__(self, gemini_api_key: str, groq_api_key: str, gemini_model: str, groq_model: str):
        self.gemini_api_key = gemini_api_key
        self.groq_api_key = groq_api_key
        self.gemini_model = gemini_model
        self.groq_model = groq_model

    def _build_analytics_context(self, request: dict) -> dict:
        """Pre-compute the numbers so the LLM can focus on generating insights."""
        receipts = request["receipts"]
        currency = request["currency"]
        total_spent = sum(r["totalAmount"] for r in receipts)

        category_totals: dict[str, float] = {}
        for receipt in receipts:
            for item in receipt["items"]:
                category = item["category"]
                category_totals[category] = category_totals.get(category, 0) + item["price"]

        sorted_categories = sorted(category_totals.items(), key=lambda x: x[1], reverse=True)

        return {
            "currency": currency,
            "total_spent": round(total_spent, 2),
            "categories": sorted_categories,
            "receipt_count": len(receipts),
            "date_range": f"{receipts[0]['date']} to {receipts[-1]['date']}",
        }

    def _build_prompt(self, context: dict) -> str:
        prompt = (
            f"You are a smart personal finance assistant helping a user understand their spending habits. "
            f"This is real receipt data from their purchases. Negative amounts represent discounts or loyalty savings.\n\n"
            f"Analyze the data below and respond with a JSON object containing exactly two fields:\n"
            f'"summary": A concise 1-2 sentence overview of their spending in {context["currency"]}. '
            f"Mention the total, the biggest spending category, and one useful observation.\n"
            f'"warnings": An array of at least 2 strings. Each warning should be a specific, actionable insight like:\n'
            f"  - Which category they're spending the most on and a practical tip to reduce it\n"
            f"  - A spending pattern they might not notice (e.g. frequent small purchases adding up)\n"
            f"  - A suggestion to set a budget for their top category\n"
            f"  - Whether they're using discounts/loyalty programs effectively\n"
            f"Do NOT flag negative amounts as errors — they are discounts. Do NOT repeat the summary in warnings.\n\n"
            f"Spending data:\n"
            f'- Period: {context["date_range"]}\n'
            f'- Total spent: {context["currency"]} {context["total_spent"]}\n'
            f'- {context["receipt_count"]} receipts\n'
            f"- Category breakdown (negative = discounts/savings):\n"
        )

        for cat, amt in context["categories"]:
            pct = round((amt / context["total_spent"]) * 100, 1)
            prompt += f"  - {cat}: {context['currency']} {round(amt, 2)} ({pct}%)\n"

        prompt += "\nRespond with ONLY the JSON object, no explanation or markdown."
        return prompt

    def _clean_llm_response(self, response_text: str) -> str:
        """Strip markdown code fences if present."""
        cleaned = response_text.strip()
        if cleaned.startswith("```"):
            cleaned = cleaned.split("\n", 1)[1] if "\n" in cleaned else cleaned[3:]
        if cleaned.endswith("```"):
            cleaned = cleaned[:-3]
        return cleaned.strip()

    def _call_gemini(self, prompt: str) -> str | None:
        if not self.gemini_api_key:
            logger.warning("GEMINI_API_KEY not set, skipping Gemini")
            return None
        try:
            client = genai.Client(api_key=self.gemini_api_key)
            response = client.models.generate_content(model=self.gemini_model, contents=prompt)
            return response.text.strip()
        except Exception as e:
            logger.error("Gemini call failed | Error=%s", str(e))
            return None

    def _call_groq(self, prompt: str) -> str | None:
        if not self.groq_api_key:
            logger.warning("GROQ_API_KEY not set, skipping Groq")
            return None
        try:
            client = Groq(api_key=self.groq_api_key)
            response = client.chat.completions.create(
                model=self.groq_model,
                messages=[{"role": "user", "content": prompt}],
                max_tokens=512,
            )
            return response.choices[0].message.content.strip()
        except Exception as e:
            logger.error("Groq call failed | Error=%s", str(e))
            return None

    def _call_llm_with_toggle(self, prompt: str) -> str | None:
        """
        Round-robin between Gemini and Groq across requests.
        If the primary provider fails, falls back to the other.
        The toggle always advances to the next provider regardless of fallback.
        """
        global _next_provider

        if _next_provider == "gemini":
            primary, fallback = self._call_gemini, self._call_groq
            primary_name, fallback_name = "Gemini", "Groq"
            _next_provider = "groq"
        else:
            primary, fallback = self._call_groq, self._call_gemini
            primary_name, fallback_name = "Groq", "Gemini"
            _next_provider = "gemini"

        logger.info("LLM toggle | Primary=%s NextTime=%s", primary_name, _next_provider)

        response_text = primary(prompt)
        if response_text:
            logger.info("LLM call succeeded | Provider=%s", primary_name)
            return response_text

        logger.warning("%s failed, falling back to %s", primary_name, fallback_name)
        response_text = fallback(prompt)
        if response_text:
            logger.info("LLM fallback succeeded | Provider=%s", fallback_name)
            return response_text

        return None

    def _ensure_min_warnings(self, warnings: list[str]) -> list[str]:
        """Ensure at least 2 warnings are present."""
        defaults = [
            "Review your top spending categories for potential savings.",
            "Consider tracking daily spending to identify patterns.",
        ]
        for default in defaults:
            if len(warnings) >= 2:
                break
            warnings.append(default)
        return warnings

    def generateInsights(self, request: dict) -> dict:
        """
        Generate natural language insights from receipt data using LLM.
        Alternates between Gemini and Groq on each call, with fallback.
        Returns response matching expense.insights.response.json schema.
        """
        context = self._build_analytics_context(request)
        prompt = self._build_prompt(context)

        logger.info("Generating insights | TotalSpent=%f ReceiptCount=%d", context["total_spent"], context["receipt_count"])

        response_text = self._call_llm_with_toggle(prompt)

        if not response_text:
            logger.error("Both LLM providers failed to generate insights")
            return {
                "summary": "Unable to generate insights.",
                "warnings": self._ensure_min_warnings([]),
            }

        cleaned = self._clean_llm_response(response_text)

        try:
            result = json.loads(cleaned)

            if "summary" not in result or not isinstance(result["summary"], str):
                result["summary"] = "Unable to generate insights."

            if "warnings" not in result or not isinstance(result["warnings"], list):
                result["warnings"] = []

            result["warnings"] = self._ensure_min_warnings(result["warnings"])

            logger.info("Insights generated successfully | WarningCount=%d", len(result["warnings"]))
            return result

        except json.JSONDecodeError:
            logger.error("Failed to parse LLM response as JSON | Response=%s", cleaned)
            return {
                "summary": cleaned,
                "warnings": self._ensure_min_warnings([]),
            }
