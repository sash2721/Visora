package models

// Frontend Request struct
type UploadRequest struct {
	Message  string `json:"message"`
	Currency string `json:"currency"`
}

// GenAI Request structs
type GenAIRequest struct {
	Image       string      `json:"image"`
	UserContext UserContext `json:"userContext"`
}

type UserContext struct {
	Currency string `json:"currency"`
	Country  string `json:"country"`
}

// GenAI Response structs
type GenAIUploadResponse struct {
	Merchant        string         `json:"merchant"`
	Date            string         `json:"date"`
	TotalAmount     float64        `json:"totalAmount"`
	Currency        string         `json:"currency"`
	Items           []ReceiptItems `json:"items"`
	ConfidenceScore float64        `json:"confidenceScore"`
}

type ReceiptItems struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// Backend Response struct
type BackendUploadResponse struct {
	ReceiptID         int                `json:"receiptID"`
	UserID            string             `json:"userID"`
	Email             string             `json:"email"`
	Merchant          string             `json:"merchant"`
	Date              string             `json:"date"`
	TotalAmount       float64            `json:"totalAmount"`
	Items             []ReceiptItems     `json:"items"`
	CategoriesSummary map[string]float64 `json:"categoriesSummary"`
}
