"""
Test file for /generatesummary endpoint logic.
Run directly: python -m tests.test_insights
This uses the LLM (Gemini/Groq) to generate natural language insights from receipt data.
Matches expense.insights.response.json schema.
"""

import json
import os
import sys
from dotenv import load_dotenv

load_dotenv(dotenv_path="../../.env")

# Add parent dir so we can import configs
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from google import genai
from groq import Groq

# Same sample data as analytics test
sample_request = {
    "userID": "user123",
    "currency": "INR",
    "period": "monthly",
    "receipts": [
        {
            "merchant": "Reliance Smart",
            "date": "2026-03-01",
            "totalAmount": 845.50,
            "items": [
                {"name": "Milk", "price": 60, "quantity": 1, "category": "Food & Groceries"},
                {"name": "Bread", "price": 45, "quantity": 1, "category": "Food & Groceries"},
                {"name": "Shampoo", "price": 240, "quantity": 2, "category": "Personal Care"},
                {"name": "Detergent", "price": 500.50, "quantity": 1, "category": "Home & Furniture"},
            ],
        },
        {
            "merchant": "Shell Petrol Pump",
            "date": "2026-03-03",
            "totalAmount": 2000,
            "items": [
                {"name": "Petrol", "price": 2000, "quantity": 1, "category": "Fuel"},
            ],
        },
        {
            "merchant": "Zara",
            "date": "2026-03-05",
            "totalAmount": 4500,
            "items": [
                {"name": "T-Shirt", "price": 1500, "quantity": 1, "category": "Clothing & Fashion"},
                {"name": "Jeans", "price": 3000, "quantity": 1, "category": "Clothing & Fashion"},
            ],
        },
        {
            "merchant": "Swiggy",
            "date": "2026-03-05",
            "totalAmount": 650,
            "items": [
                {"name": "Biryani", "price": 350, "quantity": 1, "category": "Dining Out"},
                {"name": "Butter Naan", "price": 150, "quantity": 2, "category": "Dining Out"},
                {"name": "Delivery Fee", "price": 150, "quantity": 1, "category": "Miscellaneous"},
            ],
        },
        {
            "merchant": "DMart",
            "date": "2026-03-10",
            "totalAmount": 1200,
            "items": [
                {"name": "Rice 5kg", "price": 400, "quantity": 1, "category": "Food & Groceries"},
                {"name": "Cooking Oil", "price": 250, "quantity": 1, "category": "Food & Groceries"},
                {"name": "Soap", "price": 150, "quantity": 3, "category": "Personal Care"},
                {"name": "Biscuits", "price": 400, "quantity": 4, "category": "Food & Groceries"},
            ],
        },
    ],
}


def build_analytics_context(request: dict) -> dict:
    """Pre-compute the numbers so the LLM can focus on generating insights."""
    receipts = request["receipts"]
    currency = request["currency"]
    total_spent = sum(r["totalAmount"] for r in receipts)

    category_totals: dict[str, float] = {}
    for receipt in receipts:
        for item in receipt["items"]:
            category = item["category"]
            category_totals[category] = category_totals.get(category, 0) + item["price"]

    # Sort by amount descending
    sorted_categories = sorted(category_totals.items(), key=lambda x: x[1], reverse=True)

    return {
        "currency": currency,
        "total_spent": round(total_spent, 2),
        "categories": sorted_categories,
        "receipt_count": len(receipts),
        "date_range": f"{receipts[0]['date']} to {receipts[-1]['date']}",
    }


def generate_insights(request: dict, llm_provider: str = "gemini") -> dict:
    """
    Uses LLM to generate natural language summary and warnings.
    Returns response matching expense.insights.response.json schema.
    """
    context = build_analytics_context(request)

    prompt = (
        f"You are a personal finance assistant. Analyze this spending data and respond with a JSON object "
        f"containing exactly two fields:\n"
        f'- "summary": A brief 2 or more points summary of spending in {context["currency"]}\n'
        f'- "warnings": An array of strings, each a short warning about unusual or high spending patterns. '
        f"If nothing stands out, return an empty array.\n\n"
        f"Spending data:\n"
        f'- Period: {context["date_range"]}\n'
        f'- Total spent: {context["currency"]} {context["total_spent"]}\n'
        f'- {context["receipt_count"]} receipts\n'
        f"- Category breakdown:\n"
    )

    for cat, amt in context["categories"]:
        pct = round((amt / context["total_spent"]) * 100, 1)
        prompt += f"  - {cat}: {context['currency']} {round(amt, 2)} ({pct}%)\n"

    prompt += "\nRespond with ONLY the JSON object, no explanation or markdown."

    print(f"\n--- Using {llm_provider} ---")
    print(f"Prompt:\n{prompt}\n")

    if llm_provider == "gemini":
        response_text = call_gemini(prompt)
    else:
        response_text = call_groq(prompt)

    if not response_text:
        return {"summary": "Unable to generate insights.", "warnings": []}

    # Clean up response — strip markdown code fences if present
    cleaned = response_text.strip()
    if cleaned.startswith("```"):
        cleaned = cleaned.split("\n", 1)[1] if "\n" in cleaned else cleaned[3:]
    if cleaned.endswith("```"):
        cleaned = cleaned[:-3]
    cleaned = cleaned.strip()

    try:
        result = json.loads(cleaned)
        # Validate structure
        if "summary" not in result:
            result["summary"] = "Unable to generate insights."
        if "warnings" not in result or not isinstance(result["warnings"], list):
            result["warnings"] = []
        return result
    except json.JSONDecodeError:
        print(f"Failed to parse LLM response: {cleaned}")
        return {"summary": cleaned, "warnings": []}


def call_gemini(prompt: str) -> str | None:
    api_key = os.getenv("GEMINI_API_KEY", "")
    model = os.getenv("GEMINI_MODEL", "gemini-2.5-flash")
    if not api_key:
        print("GEMINI_API_KEY not set, skipping Gemini")
        return None
    try:
        client = genai.Client(api_key=api_key)
        response = client.models.generate_content(model=model, contents=prompt)
        return response.text.strip()
    except Exception as e:
        print(f"Gemini error: {e}")
        return None


def call_groq(prompt: str) -> str | None:
    api_key = os.getenv("GROQ_API_KEY", "")
    model = os.getenv("GROQ_MODEL", "llama-3.3-70b-versatile")
    if not api_key:
        print("GROQ_API_KEY not set, skipping Groq")
        return None
    try:
        client = Groq(api_key=api_key)
        response = client.chat.completions.create(
            model=model,
            messages=[{"role": "user", "content": prompt}],
            max_tokens=512,
        )
        return response.choices[0].message.content.strip()
    except Exception as e:
        print(f"Groq error: {e}")
        return None


if __name__ == "__main__":
    print("=" * 50)
    print("Testing Insights Generation")
    print("=" * 50)

    # Test with Gemini
    gemini_result = generate_insights(sample_request, llm_provider="gemini")
    print("\n=== Gemini Response ===")
    print(json.dumps(gemini_result, indent=2))

    # Test with Groq
    groq_result = generate_insights(sample_request, llm_provider="groq")
    print("\n=== Groq Response ===")
    print(json.dumps(groq_result, indent=2))

    # Validate structure for whichever succeeded
    for name, result in [("Gemini", gemini_result), ("Groq", groq_result)]:
        assert "summary" in result, f"{name}: missing 'summary'"
        assert "warnings" in result, f"{name}: missing 'warnings'"
        assert isinstance(result["summary"], str), f"{name}: 'summary' should be a string"
        assert isinstance(result["warnings"], list), f"{name}: 'warnings' should be a list"
        print(f"✅ {name} response structure valid!")
