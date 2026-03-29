from pydantic import BaseModel


# Request models
class UserContext(BaseModel):
    currency: str
    country: str


class UploadReceiptRequest(BaseModel):
    image: str  # base64 encoded image
    userContext: UserContext


# Response models
class ReceiptItem(BaseModel):
    name: str
    price: float
    category: str


class UploadReceiptResponse(BaseModel):
    merchant: str
    date: str
    totalAmount: float
    currency: str
    items: list[ReceiptItem]
    confidenceScore: float
