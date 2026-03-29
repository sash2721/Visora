package handlers

import (
	"Backend/errors"
	"Backend/models"
	"Backend/services"
	"encoding/json"
	"log/slog"
	"net/http"
)

type UploadHandler struct {
	Service *services.UploadService
}

func (s *UploadHandler) HandleReceiptUploads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Only POST method allowed"}`))
		return
	}

	var uploadReq models.UploadRequest
	err := json.NewDecoder(r.Body).Decode(&uploadReq)

	if err != nil {
		slog.Debug(
			"Incorrect request object sent",
		)

		errorJson, badRequestError := errors.NewBadRequestError("Incorrect request object sent", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(badRequestError.Code)
		w.Write(errorJson)
		return
	}

	// extracting user info from the request context (set by AuthZ middleware)
	userID := r.Context().Value("userID").(string)
	email := r.Context().Value("email").(string)

	// sending the request data to the service
	responseData, err, errCode, errJsonData := s.Service.ProcessReceiptImage(uploadReq.Message, uploadReq.Currency, userID, email, r)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errCode)
		w.Write(errJsonData)

		slog.Error(
			"Receipt image upload failed!",
			slog.Any("Error", err),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)

	slog.Info("Receipt image uploaded successfully!", slog.String("UserID", userID))
}
