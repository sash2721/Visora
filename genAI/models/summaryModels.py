from pydantic import BaseModel


# Shared request model for analytics and insights
class ReceiptItem(BaseModel):
    name: str
    price: float
    quantity: int = 1
    category: str


class UserReceipt(BaseModel):
    merchant: str
    date: str
    totalAmount: float
    items: list[ReceiptItem]


class SummaryRequest(BaseModel):
    userID: str
    currency: str
    period: str
    receipts: list[UserReceipt]


# Analytics response models
class CategoryItem(BaseModel):
    category: str
    amount: float


class DailySpending(BaseModel):
    date: str
    amount: float


class GetAnalyticsResponse(BaseModel):
    totalAmount: float
    categoryBreakdown: list[CategoryItem]
    dailySpending: list[DailySpending]


# Insights response model
class GetInsightsResponse(BaseModel):
    summary: str
    warnings: list[str]
