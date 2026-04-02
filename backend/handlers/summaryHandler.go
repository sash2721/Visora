package handlers

import (
	"Backend/errors"
	"Backend/services"
	"encoding/json"
	"log/slog"
	"net/http"
)

type SummaryHandler struct {
	Service *services.SummaryService
}

func (h *SummaryHandler) HandleGetAnalytics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	slog.Info("Get analytics request received", slog.String("UserID", userID))

	analytics, err := h.Service.GetCachedAnalytics(userID)
	if err != nil {
		slog.Error("Failed to get cached analytics", slog.String("UserID", userID), slog.Any("Error", err))
		errJSON, internalErr := errors.NewInternalServerError("Error while fetching analytics data", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(internalErr.Code)
		w.Write(errJSON)
		return
	}

	if analytics == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "No analytics data available yet. Upload a receipt to get started."}`))
		return
	}

	jsonResponse, err := json.Marshal(analytics)
	if err != nil {
		slog.Error("Failed to marshal analytics response", slog.String("UserID", userID), slog.Any("Error", err))
		errJSON, internalErr := errors.NewInternalServerError("Error while preparing analytics response", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(internalErr.Code)
		w.Write(errJSON)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (h *SummaryHandler) HandleGetInsights(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	slog.Info("Get insights request received", slog.String("UserID", userID))

	insights, err := h.Service.GetCachedInsights(userID)
	if err != nil {
		slog.Error("Failed to get cached insights", slog.String("UserID", userID), slog.Any("Error", err))
		errJSON, internalErr := errors.NewInternalServerError("Error while fetching insights data", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(internalErr.Code)
		w.Write(errJSON)
		return
	}

	if insights == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "No insights data available yet. Upload a receipt to get started."}`))
		return
	}

	jsonResponse, err := json.Marshal(insights)
	if err != nil {
		slog.Error("Failed to marshal insights response", slog.String("UserID", userID), slog.Any("Error", err))
		errJSON, internalErr := errors.NewInternalServerError("Error while preparing insights response", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(internalErr.Code)
		w.Write(errJSON)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
