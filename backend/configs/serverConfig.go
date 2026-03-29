package configs

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	BackendPort         string
	BackendHost         string
	BackendLoginAPI     string
	BackendSignupAPI    string
	BackendUploadAPI    string
	GenAIPort           string
	GenAIHost           string
	GenAIUploadEndpoint string
	Env                 string
	SecretKey           string
	OpenAIAPIKey        string
}

func GetServerConfig() *ServerConfig {
	err := godotenv.Load("../.env")

	if err != nil {
		slog.Warn(
			"Error loading .env file, using system environment variables",
		)
	}

	serverConfig := &ServerConfig{
		BackendPort:         os.Getenv("BACKEND_PORT"),
		BackendHost:         os.Getenv("BACKEND_HOST"),
		BackendLoginAPI:     os.Getenv("BACKEND_LOGIN_API"),
		BackendSignupAPI:    os.Getenv("BACKEND_SIGNUP_API"),
		BackendUploadAPI:    os.Getenv("BACKEND_UPLOAD_API"),
		GenAIPort:           os.Getenv("GENAI_PORT"),
		GenAIHost:           os.Getenv("GENAI_HOST"),
		GenAIUploadEndpoint: os.Getenv("GENAI_UPLOAD_API"),
		Env:                 os.Getenv("ENV"),
		SecretKey:           os.Getenv("ENCRYPTION_SECRET_KEY"),
		OpenAIAPIKey:        os.Getenv("OPENAI_API_KEY"),
	}

	return serverConfig
}
