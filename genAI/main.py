from fastapi import FastAPI
from models.uploadModels import UploadReceiptRequest, UploadReceiptResponse
from configs.serverConfig import ServerConfig
from cv.processReceipt import ProcessReceipts

app = FastAPI()
config = ServerConfig()

print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
print("Starting GenAI Service")
print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")


@app.get("/health")
def read_health():
    return {"message": "GenAI Service running"}


@app.post(config.GENAI_UPLOAD_API, response_model=UploadReceiptResponse)
def process_uploaded_receipt(request: UploadReceiptRequest):
    # initialising the CV Service handler
    cvHandler = ProcessReceipts()
    receiptData: dict = cvHandler.convertImageToData(request.image, request.userContext.currency)

    return receiptData

@app.get(config.GENAI_GENERATE_SUMMARY_API)
def generate_llm_summary():
    # TODO: For later
    pass

@app.get(config.GENAI_GET_ANALYTICS_API)
def get_user_analytics():
    # TODO: For later
    pass