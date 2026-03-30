import os
from dotenv import load_dotenv

load_dotenv(dotenv_path="../.env")


class ServerConfig:
    GENAI_HOST = os.getenv("GENAI_HOST", "localhost")
    GENAI_PORT = os.getenv("GENAI_PORT", ":4000")
    GENAI_UPLOAD_API = os.getenv("GENAI_UPLOAD_API", "/uploadreceipt")
    GENAI_GENERATE_SUMMARY_API = os.getenv("GENAI_GENERATE_SUMMARY_API", "/generatesummary")
    GENAI_GET_ANALYTICS_API = os.getenv("GENAI_GET_ANALYTICS_API", "/getanalytics")
    GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
    GROQ_API_KEY = os.getenv("GROQ_API_KEY", "")
    OCR_API_KEY = os.getenv("OCR_API_KEY", "")
    MODEL_ID = os.getenv("OCR_MODEL_ID")
