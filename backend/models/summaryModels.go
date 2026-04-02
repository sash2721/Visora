package models

import "encoding/json"

// --- GenAI Request structs ---

type GenAISummaryReceiptItem struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Category string  `json:"category"`
}

type GenAISummaryReceipt struct {
	Merchant    string                    `json:"merchant"`
	Date        string                    `json:"date"`
	TotalAmount float64                   `json:"totalAmount"`
	Items       []GenAISummaryReceiptItem `json:"items"`
}

type GenAISummaryRequest struct {
	UserID   string                `json:"userID"`
	Currency string                `json:"currency"`
	Period   string                `json:"period"`
	Receipts []GenAISummaryReceipt `json:"receipts"`
}

// --- GenAI Response structs ---

type CategoryBreakdownItem struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

type DailySpendingItem struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type GenAIAnalyticsResponse struct {
	TotalAmount       float64                 `json:"totalAmount"`
	CategoryBreakdown []CategoryBreakdownItem `json:"categoryBreakdown"`
	DailySpending     []DailySpendingItem     `json:"dailySpending"`
}

type GenAIInsightsResponse struct {
	Summary  string   `json:"summary"`
	Warnings []string `json:"warnings"`
}

// --- Backend Response structs (served to frontend) ---

type CachedAnalyticsResponse struct {
	TotalSpent        float64         `json:"totalSpent"`
	Currency          string          `json:"currency"`
	CategoryBreakdown json.RawMessage `json:"categoryBreakdown"`
	DailySpending     json.RawMessage `json:"dailySpending"`
	Period            string          `json:"period"`
	ComputedAt        string          `json:"computedAt"`
}

type CachedInsightsResponse struct {
	Summary    string          `json:"summary"`
	Warnings   json.RawMessage `json:"warnings"`
	Period     string          `json:"period"`
	ComputedAt string          `json:"computedAt"`
}
