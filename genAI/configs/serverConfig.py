import os
from dotenv import load_dotenv

load_dotenv(dotenv_path="../.env")


class ServerConfig:
    GENAI_HOST = os.getenv("GENAI_HOST", "localhost")
    GENAI_PORT = os.getenv("GENAI_PORT", ":4000")
    GENAI_UPLOAD_API = os.getenv("GENAI_UPLOAD_API", "/uploadreceipt")
    GENAI_GENERATE_SUMMARY_API = os.getenv("GENAI_GENERATE_SUMMARY_API", "/generatesummary")
    GENAI_GET_ANALYTICS_API = os.getenv("GENAI_GET_ANALYTICS_API", "/getanalytics")
    OPENAI_API_KEY = os.getenv("OPENAI_API_KEY", "")
