package services

import (
	"Backend/configs"
	"Backend/models"
	"Backend/repositories"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type SummaryService struct {
	Repo *repositories.SummaryRepository
}

func NewSummaryService(repo *repositories.SummaryRepository) *SummaryService {
	return &SummaryService{Repo: repo}
}

func buildGenAIPayload(userID string, rows []repositories.ReceiptRow) models.GenAISummaryRequest {
	receiptsMap := make(map[string]*models.GenAISummaryReceipt)
	var orderedIDs []string
	var currency string

	for _, row := range rows {
		if currency == "" {
			currency = row.Currency
		}

		if _, exists := receiptsMap[row.ReceiptID]; !exists {
			receiptsMap[row.ReceiptID] = &models.GenAISummaryReceipt{
				Merchant:    row.Merchant,
				Date:        row.Date.Format("2006-01-02"),
				TotalAmount: row.TotalAmount,
				Items:       []models.GenAISummaryReceiptItem{},
			}
			orderedIDs = append(orderedIDs, row.ReceiptID)
		}

		receiptsMap[row.ReceiptID].Items = append(receiptsMap[row.ReceiptID].Items, models.GenAISummaryReceiptItem{
			Name:     row.ItemName,
			Price:    row.ItemPrice,
			Quantity: row.ItemQuantity,
			Category: row.CategoryName,
		})
	}

	receipts := make([]models.GenAISummaryReceipt, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		receipts = append(receipts, *receiptsMap[id])
	}

	return models.GenAISummaryRequest{
		UserID:   userID,
		Currency: currency,
		Period:   "monthly",
		Receipts: receipts,
	}
}

func callGenAI(endpoint string, payload any) ([]byte, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal GenAI request body", slog.String("Endpoint", endpoint))
		return nil, fmt.Errorf("error while forming the request body to send to GenAI service: %w", err)
	}

	slog.Info("Sending request to GenAI service",
		slog.String("URL", endpoint),
		slog.Int("PayloadSizeBytes", len(jsonBody)),
	)

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		slog.Error("Failed to send request to GenAI service", slog.String("URL", endpoint))
		return nil, fmt.Errorf("error while sending the request to GenAI service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read GenAI response body", slog.String("URL", endpoint))
		return nil, fmt.Errorf("error while reading the response from GenAI service: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("GenAI service returned error",
			slog.Int("StatusCode", resp.StatusCode),
			slog.Int("ResponseSizeBytes", len(body)),
		)
		return nil, fmt.Errorf("GenAI service error (status %d): %s", resp.StatusCode, string(body))
	}

	slog.Info("GenAI response received",
		slog.Int("StatusCode", resp.StatusCode),
		slog.Int("ResponseSizeBytes", len(body)),
	)

	return body, nil
}

func (s *SummaryService) RecomputeAnalytics(userID string) error {
	slog.Info("Recomputing analytics", slog.String("UserID", userID))

	rows, err := s.Repo.GetUserReceipts(userID)
	if err != nil {
		slog.Error("Failed to fetch receipts from DB", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while fetching receipt data for analytics: %w", err)
	}

	if len(rows) == 0 {
		slog.Info("No receipts found, skipping analytics recomputation", slog.String("UserID", userID))
		return nil
	}

	payload := buildGenAIPayload(userID, rows)

	cfg := configs.GetServerConfig()
	url := fmt.Sprintf("http://%s%s%s", cfg.GenAIHost, cfg.GenAIPort, cfg.GenAIGetAnalyticsEndpoint)

	body, err := callGenAI(url, payload)
	if err != nil {
		slog.Error("GenAI analytics call failed", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while computing analytics via GenAI: %w", err)
	}

	var analytics models.GenAIAnalyticsResponse
	if err := json.Unmarshal(body, &analytics); err != nil {
		slog.Error("Failed to unmarshal GenAI analytics response", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while unmarshaling the GenAI analytics response: %w", err)
	}

	categoryJSON, err := json.Marshal(analytics.CategoryBreakdown)
	if err != nil {
		slog.Error("Failed to marshal category breakdown", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while marshaling category breakdown: %w", err)
	}

	dailyJSON, err := json.Marshal(analytics.DailySpending)
	if err != nil {
		slog.Error("Failed to marshal daily spending", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while marshaling daily spending: %w", err)
	}

	if err := s.Repo.UpsertAnalytics(userID, analytics.TotalAmount, payload.Currency, categoryJSON, dailyJSON, "monthly"); err != nil {
		slog.Error("Failed to upsert analytics into DB", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while storing analytics data: %w", err)
	}

	slog.Info("Analytics recomputed and cached successfully", slog.String("UserID", userID))
	return nil
}

func (s *SummaryService) RecomputeInsights(userID string) error {
	slog.Info("Recomputing insights", slog.String("UserID", userID))

	rows, err := s.Repo.GetUserReceipts(userID)
	if err != nil {
		slog.Error("Failed to fetch receipts from DB", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while fetching receipt data for insights: %w", err)
	}

	if len(rows) == 0 {
		slog.Info("No receipts found, skipping insights recomputation", slog.String("UserID", userID))
		return nil
	}

	payload := buildGenAIPayload(userID, rows)

	cfg := configs.GetServerConfig()
	url := fmt.Sprintf("http://%s%s%s", cfg.GenAIHost, cfg.GenAIPort, cfg.GenAIGenerateSummaryEndpoint)

	body, err := callGenAI(url, payload)
	if err != nil {
		slog.Error("GenAI insights call failed", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while generating insights via GenAI: %w", err)
	}

	var insights models.GenAIInsightsResponse
	if err := json.Unmarshal(body, &insights); err != nil {
		slog.Error("Failed to unmarshal GenAI insights response", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while unmarshaling the GenAI insights response: %w", err)
	}

	warningsJSON, err := json.Marshal(insights.Warnings)
	if err != nil {
		slog.Error("Failed to marshal warnings", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while marshaling insights warnings: %w", err)
	}

	if err := s.Repo.UpsertInsights(userID, insights.Summary, warningsJSON, "monthly"); err != nil {
		slog.Error("Failed to upsert insights into DB", slog.String("UserID", userID), slog.Any("Error", err))
		return fmt.Errorf("error while storing insights data: %w", err)
	}

	slog.Info("Insights recomputed and cached successfully", slog.String("UserID", userID))
	return nil
}

func (s *SummaryService) GetCachedAnalytics(userID string) (*models.CachedAnalyticsResponse, error) {
	slog.Info("Fetching cached analytics", slog.String("UserID", userID))

	cached, err := s.Repo.GetCachedAnalytics(userID, "monthly")
	if err != nil {
		slog.Error("Failed to fetch cached analytics from DB", slog.String("UserID", userID), slog.Any("Error", err))
		return nil, fmt.Errorf("error while fetching cached analytics: %w", err)
	}

	if cached == nil {
		slog.Info("No cached analytics found", slog.String("UserID", userID))
		return nil, nil
	}

	return &models.CachedAnalyticsResponse{
		TotalSpent:        cached.TotalSpent,
		Currency:          cached.Currency,
		CategoryBreakdown: cached.CategoryBreakdown,
		DailySpending:     cached.DailySpending,
		Period:            cached.Period,
		ComputedAt:        cached.ComputedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *SummaryService) GetCachedInsights(userID string) (*models.CachedInsightsResponse, error) {
	slog.Info("Fetching cached insights", slog.String("UserID", userID))

	cached, err := s.Repo.GetCachedInsights(userID, "monthly")
	if err != nil {
		slog.Error("Failed to fetch cached insights from DB", slog.String("UserID", userID), slog.Any("Error", err))
		return nil, fmt.Errorf("error while fetching cached insights: %w", err)
	}

	if cached == nil {
		slog.Info("No cached insights found", slog.String("UserID", userID))
		return nil, nil
	}

	return &models.CachedInsightsResponse{
		Summary:    cached.Summary,
		Warnings:   cached.Warnings,
		Period:     cached.Period,
		ComputedAt: cached.ComputedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
