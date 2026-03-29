package services

import (
	"Backend/configs"
	"Backend/errors"
	"Backend/models"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
)

type UploadService struct{}

var serverConfig configs.ServerConfig

var GenAIURL string = fmt.Sprintf("http://%s:%s/%s", serverConfig.GenAIHost, serverConfig.GenAIPort, serverConfig.GenAIUploadEndpoint)

func NewUploadService() *UploadService {
	return &UploadService{}
}

func (*UploadService) ProcessReceiptImage(message string, currency string, userID string, email string, r *http.Request) ([]byte, error, int, []byte) {
	// parsing the multipart form
	err := r.ParseMultipartForm(10 << 20)

	if err != nil {
		slog.Debug(
			"Error while parsing the file, file too large or bad request",
			slog.Any("Error", err),
			slog.Any("Request Body", r.Body),
		)
		errJsonData, badRequestError := errors.NewBadRequestError("Error while parsing the file, file too large or bad request!", err)
		return nil, badRequestError, badRequestError.Code, errJsonData
	}

	// extract the image from the file field
	file, _, err := r.FormFile("image")

	if err != nil {
		slog.Error(
			"Error while extracting the uploaded receipt image, bad request!",
			slog.Any("Error", err),
			slog.Any("Request Body", r.Body),
		)
		errJsonData, badRequestError := errors.NewBadRequestError("Error while extracting the uploaded receipt image, bad request!", err)
		return nil, badRequestError, badRequestError.Code, errJsonData
	}

	defer file.Close() // close the file at the end of all the processing

	// send the request to GenAI service
	responseData, responseError, responseErrorCode, errorData := sendUploadReceiptToGenAI(file, currency)

	if responseError != nil {
		return nil, responseError, responseErrorCode, errorData
	}

	var uploadResp models.GenAIUploadResponse
	err = json.Unmarshal(responseData, &uploadResp)

	if err != nil {
		slog.Error(
			"Error while unmarshaling the GenAI response!",
			slog.Any("Error", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while unmarshaling the GenAI response!", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	// TODO: store this data in the DB and get the receipt ID

	// building the backend response
	var receiptID int = 0
	var merchant string = uploadResp.Merchant
	var date string = uploadResp.Date
	var totalAmount float64 = uploadResp.TotalAmount
	var receiptItems []models.ReceiptItems = uploadResp.Items
	var categoriesSummary map[string]float64 = buildCategoriesSummary(receiptItems)

	backendResp := models.BackendUploadResponse{
		ReceiptID:         receiptID,
		UserID:            userID,
		Email:             email,
		Merchant:          merchant,
		Date:              date,
		TotalAmount:       totalAmount,
		Items:             receiptItems,
		CategoriesSummary: categoriesSummary,
	}

	// marshaling the response in json
	jsonResponse, err := json.Marshal(backendResp)

	if err != nil {
		slog.Error(
			"Error while marshaling the backend response to send to frontend!",
			slog.Any("Error", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while marshaling the backend response to send to frontend!", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	return jsonResponse, nil, 0, nil
}

func sendUploadReceiptToGenAI(receiptImage multipart.File, currency string) ([]byte, error, int, []byte) {
	// read the file bytes
	imageBytes, err := io.ReadAll(receiptImage)

	if err != nil {
		slog.Error(
			"Error while reading the image bytes!",
			slog.Any("Error", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while reading the image bytes!", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	genAIReq := models.GenAIRequest{
		Image: base64.StdEncoding.EncodeToString(imageBytes),
		UserContext: models.UserContext{
			Currency: currency,
			Country:  "",
		},
	}

	// marshal the struct to JSON before sending
	jsonBody, err := json.Marshal(genAIReq)

	if err != nil {
		slog.Error(
			"Error while forming the request body to send to GenAI service",
			slog.Any("Error", err),
			slog.Any("Body", jsonBody),
		)
	}

	// make the POST request
	response, err := http.Post(GenAIURL, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		slog.Error(
			"Error while sending the request to Gen AI service",
			slog.Any("Error", err),
			slog.Any("Request", jsonBody),
		)

		errJsonData, internalServerError := errors.NewInternalServerError("Error while sending the request to Gen AI service", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	defer response.Body.Close()

	// read the reponse
	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		slog.Error(
			"Error while reading the response for Upload request",
			slog.Any("Error", err),
			slog.Any("Request", jsonBody),
		)

		errJsonData, internalServerError := errors.NewInternalServerError("Error while reading the response for Upload request", err)
		return nil, internalServerError, internalServerError.Code, errJsonData
	}

	// return the successful response
	return responseBody, nil, 0, nil
}

func buildCategoriesSummary(itemsList []models.ReceiptItems) map[string]float64 {
	summary := make(map[string]float64)
	for _, item := range itemsList {
		summary[item.Category] += item.Price
	}
	return summary
}
