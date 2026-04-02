package configs

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	BackendPort                  string
	BackendHost                  string
	BackendLoginAPI              string
	BackendSignupAPI             string
	BackendUploadAPI             string
	BackendManualExpenseAPI      string
	BackendAnalyticsAPI          string
	BackendInsightsAPI           string
	GenAIPort                    string
	GenAIHost                    string
	GenAIUploadEndpoint          string
	GenAIGetAnalyticsEndpoint    string
	GenAIGenerateSummaryEndpoint string
	Env                          string
	SecretKey                    string
	DatabaseConnectionString     string
	BucketEndpoint               string
}

func GetServerConfig() *ServerConfig {
	err := godotenv.Load("../.env")

	if err != nil {
		slog.Warn(
			"Error loading .env file, using system environment variables",
		)
	}

	serverConfig := &ServerConfig{
		BackendPort:                  os.Getenv("BACKEND_PORT"),
		BackendHost:                  os.Getenv("BACKEND_HOST"),
		BackendLoginAPI:              os.Getenv("BACKEND_LOGIN_API"),
		BackendSignupAPI:             os.Getenv("BACKEND_SIGNUP_API"),
		BackendUploadAPI:             os.Getenv("BACKEND_UPLOAD_API"),
		BackendManualExpenseAPI:      os.Getenv("BACKEND_MANUAL_EXPENSE_API"),
		BackendAnalyticsAPI:          os.Getenv("BACKEND_ANALYTICS_API"),
		BackendInsightsAPI:           os.Getenv("BACKEND_INSIGHTS_API"),
		GenAIPort:                    os.Getenv("GENAI_PORT"),
		GenAIHost:                    os.Getenv("GENAI_HOST"),
		GenAIUploadEndpoint:          os.Getenv("GENAI_UPLOAD_API"),
		GenAIGetAnalyticsEndpoint:    os.Getenv("GENAI_GET_ANALYTICS_API"),
		GenAIGenerateSummaryEndpoint: os.Getenv("GENAI_GENERATE_SUMMARY_API"),
		Env:                          os.Getenv("ENV"),
		SecretKey:                    os.Getenv("ENCRYPTION_SECRET_KEY"),
		DatabaseConnectionString:     os.Getenv("DATABASE_CONNECTION_STRING"),
		BucketEndpoint:               os.Getenv("BUCKET_ENDPOINT_STRING"),
	}

	return serverConfig
}
