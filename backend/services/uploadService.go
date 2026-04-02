package services

import (
	"Backend/configs"
	"Backend/errors"
	"Backend/models"
	"Backend/repositories"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
)

type UploadService struct {
	Repo           *repositories.UploadRepository
	SummaryService *SummaryService
}

var serverConfig *configs.ServerConfig

func init() {
	serverConfig = configs.GetServerConfig()
}

func getGenAIURL() string {
	return fmt.Sprintf("http://%s%s%s", serverConfig.GenAIHost, serverConfig.GenAIPort, serverConfig.GenAIUploadEndpoint)
}

func NewUploadService(repo *repositories.UploadRepository, summaryService *SummaryService) *UploadService {
	return &UploadService{Repo: repo, SummaryService: summaryService}
}

func (s *UploadService) ProcessReceiptImage(file multipart.File, currency string, userID string, email string) ([]byte, error, int, []byte) {
	slog.Info("Processing receipt image",
		slog.String("UserID", userID),
		slog.String("Currency", currency),
	)

	responseData, responseError, responseErrorCode, errorData := sendUploadReceiptToGenAI(file, currency)

	if responseError != nil {
		return nil, responseError, responseErrorCode, errorData
	}

	var uploadResp models.GenAIUploadResponse
	err := json.Unmarshal(responseData, &uploadResp)

	if err != nil {
		slog.Error("Failed to unmarshal GenAI response")
		errJsonData, internalServerError := errors.NewInternalServerError("Error while unmarshaling the GenAI response!", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	slog.Info("GenAI response parsed successfully",
		slog.String("Merchant", uploadResp.Merchant),
		slog.String("Date", uploadResp.Date),
		slog.Float64("TotalAmount", uploadResp.TotalAmount),
		slog.String("Currency", uploadResp.Currency),
		slog.Int("ItemCount", len(uploadResp.Items)),
		slog.Float64("ConfidenceScore", uploadResp.ConfidenceScore),
	)

	// storing the receipt data in the database
	receiptID, err := s.Repo.StoreReceipt(userID, uploadResp, "")
	if err != nil {
		slog.Error("Failed to store receipt in DB", slog.Any("Error", err))
		errJsonData, internalServerError := errors.NewInternalServerError("Error while storing receipt data", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	var merchant string = uploadResp.Merchant
	var date string = uploadResp.Date
	var totalAmount float64 = uploadResp.TotalAmount
	var receiptCurrency string = uploadResp.Currency
	var receiptItems []models.ReceiptItems = uploadResp.Items
	var categoriesSummary map[string]float64 = buildCategoriesSummary(receiptItems)

	backendResp := models.BackendUploadResponse{
		ReceiptID:         receiptID,
		UserID:            userID,
		Email:             email,
		Merchant:          merchant,
		Date:              date,
		TotalAmount:       totalAmount,
		Currency:          receiptCurrency,
		Items:             receiptItems,
		CategoriesSummary: categoriesSummary,
		ConfidenceScore:   uploadResp.ConfidenceScore,
	}

	jsonResponse, err := json.Marshal(backendResp)

	if err != nil {
		slog.Error("Failed to marshal backend response")
		errJsonData, internalServerError := errors.NewInternalServerError("Error while marshaling the backend response to send to frontend!", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	slog.Info("Backend response built successfully",
		slog.String("UserID", userID),
		slog.Int("CategoryCount", len(categoriesSummary)),
	)

	// recompute analytics & insights in the background after successful upload
	go func() {
		if err := s.SummaryService.RecomputeAnalytics(userID); err != nil {
			slog.Error("Failed to recompute analytics", slog.String("UserID", userID), slog.Any("Error", err))
		}
		if err := s.SummaryService.RecomputeInsights(userID); err != nil {
			slog.Error("Failed to recompute insights", slog.String("UserID", userID), slog.Any("Error", err))
		}
	}()

	return jsonResponse, nil, 0, nil
}

func sendUploadReceiptToGenAI(receiptImage multipart.File, currency string) ([]byte, error, int, []byte) {
	imageBytes, err := io.ReadAll(receiptImage)

	if err != nil {
		slog.Error("Failed to read image bytes")
		errJsonData, internalServerError := errors.NewInternalServerError("Error while reading the image bytes!", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	slog.Info("Image read successfully",
		slog.Int("ImageSizeBytes", len(imageBytes)),
	)

	genAIReq := models.GenAIRequest{
		Image: base64.StdEncoding.EncodeToString(imageBytes),
		UserContext: models.UserContext{
			Currency: currency,
			Country:  "",
		},
	}

	jsonBody, err := json.Marshal(genAIReq)

	if err != nil {
		slog.Error("Failed to marshal GenAI request body")
		errJsonData, internalServerError := errors.NewInternalServerError("Error while forming the request body to send to GenAI service", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	genAIURL := getGenAIURL()
	slog.Info("Sending request to GenAI service",
		slog.String("URL", genAIURL),
		slog.Int("PayloadSizeBytes", len(jsonBody)),
	)

	response, err := http.Post(genAIURL, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		slog.Error("Failed to send request to GenAI service",
			slog.String("URL", genAIURL),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while sending the request to Gen AI service", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		slog.Error("Failed to read GenAI response body")
		errJsonData, internalServerError := errors.NewInternalServerError("Error while reading the response for Upload request", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	if response.StatusCode != http.StatusOK {
		slog.Error("GenAI service returned error",
			slog.Int("StatusCode", response.StatusCode),
			slog.Int("ResponseSizeBytes", len(responseBody)),
		)
		errJsonData, internalServerError := errors.NewInternalServerError(
			fmt.Sprintf("GenAI service error (status %d)", response.StatusCode), nil)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	slog.Info("GenAI response received",
		slog.Int("StatusCode", response.StatusCode),
		slog.Int("ResponseSizeBytes", len(responseBody)),
	)

	return responseBody, nil, 0, nil
}

func buildCategoriesSummary(itemsList []models.ReceiptItems) map[string]float64 {
	summary := make(map[string]float64)
	for _, item := range itemsList {
		summary[item.Category] += item.Price
	}
	return summary
}

func (s *UploadService) ProcessManualExpense(expense models.ManualExpenseRequest, userID string, email string) ([]byte, error, int, []byte) {
	slog.Info("Processing manual expense entry",
		slog.String("UserID", userID),
		slog.String("Merchant", expense.Merchant),
	)

	// calculate total from items
	var totalAmount float64
	for _, item := range expense.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	receiptID, err := s.Repo.StoreManualExpense(userID, expense, totalAmount)
	if err != nil {
		slog.Error("Failed to store manual expense in DB", slog.Any("Error", err))
		errJsonData, internalServerError := errors.NewInternalServerError("Error while storing manual expense data", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	categoriesSummary := buildCategoriesSummary(expense.Items)

	backendResp := models.BackendUploadResponse{
		ReceiptID:         receiptID,
		UserID:            userID,
		Email:             email,
		Merchant:          expense.Merchant,
		Date:              expense.Date,
		TotalAmount:       totalAmount,
		Currency:          expense.Currency,
		Items:             expense.Items,
		CategoriesSummary: categoriesSummary,
		ConfidenceScore:   1.0,
	}

	jsonResponse, err := json.Marshal(backendResp)
	if err != nil {
		slog.Error("Failed to marshal manual expense response")
		errJsonData, internalServerError := errors.NewInternalServerError("Error while marshaling the manual expense response", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	// recompute analytics & insights in the background
	go func() {
		if err := s.SummaryService.RecomputeAnalytics(userID); err != nil {
			slog.Error("Failed to recompute analytics", slog.String("UserID", userID), slog.Any("Error", err))
		}
		if err := s.SummaryService.RecomputeInsights(userID); err != nil {
			slog.Error("Failed to recompute insights", slog.String("UserID", userID), slog.Any("Error", err))
		}
	}()

	slog.Info("Manual expense processed successfully",
		slog.String("UserID", userID),
		slog.String("ReceiptID", receiptID),
	)

	return jsonResponse, nil, 0, nil
}
