package services

import (
	"Backend/errors"
	"Backend/utils"
	"log/slog"
	"net/http"
	"strings"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// imitating the repository layer for now
var passwordCache = map[string]string{}

var adminEmails = []string{
	"sahilshah2104@gmail.com",
	"sakshipaygude27@gmail.com",
}

func (*AuthService) Login(email string, password string) (string, string, string, string, error, []byte, int) {
	// get the hashed password from the passwordCache
	hashedPassword, exists := passwordCache[email]

	if !exists {
		slog.Debug(
			"User not present please login!",
			slog.String("Email", email),
		)
		errorJson, badRequestError := errors.NewBadRequestError("User not present please signup!", nil)
		return "", "", "", "", badRequestError, errorJson, badRequestError.Code
	}

	// send this hashedPassword and original password for comparison
	err := utils.ComparePassword(hashedPassword, password)

	if err != nil {
		slog.Error(
			"Password is incorrect, please retry",
			slog.Any("Error", err),
		)
		errorJson, badRequestError := errors.NewBadRequestError("Password is incorrect, please retry", err)
		return "", "", "", "", badRequestError, errorJson, badRequestError.Code
	}

	// extracting the userID from the mail itself
	userID := strings.Split(email, "@")[0]

	// Getting the user role by checking the email in db
	var role string = "user"
	for _, adminEmail := range adminEmails {
		if email == adminEmail {
			role = "admin"
			break
		}
	}

	// generating the JWT token for this part
	jwtToken, err := utils.GenerateToken(userID, email, role)

	if err != nil {
		slog.Error(
			"Error while generating the JWT token",
			slog.Any("Error", err),
		)
		errorJson, internalServerError := errors.NewInternalServerError("Error while generating the JWT token", err)
		return "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	return jwtToken, userID, email, role, nil, nil, 0
}

func (*AuthService) Signup(email string, password string) (string, string, string, string, error, []byte, int) {
	// Check if user already exists
	if _, exists := passwordCache[email]; exists {
		slog.Debug("User already exists!", slog.String("Email", email))
		errorJson, badRequestError := errors.NewBadRequestError("User already exists, please login instead", nil)
		return "", "", "", "", badRequestError, errorJson, badRequestError.Code
	}

	// Hash the incoming password
	hashedPassword, err := utils.HashPassword(password)

	if err != nil {
		slog.Error(
			"Error while hashing the password",
			slog.Any("Error", err),
		)
		errorJson, internalServerError := errors.NewInternalServerError("Error while hashing the password", err)
		return "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	slog.Debug("Successfully hashed the incoming password", slog.String("Email", email))

	// store the hashedpassword along with the mail in the database (imitating the db for now)
	passwordCache[email] = hashedPassword

	slog.Debug("Retrieved the user hashed password from the database (repository)", slog.String("Email", email))

	// once the signup process is completed then autologin
	jwttoken, userID, userEmail, role, err, errJson, errorCode := NewAuthService().Login(email, password)

	if err != nil {
		if errorCode == http.StatusBadRequest {
			slog.Debug(
				"Invalid credentials! Either email or password is incorrect",
				slog.String("Email", email),
				slog.Any("Error", err),
			)
			return "", "", "", "", err, errJson, errorCode
		} else if errorCode == http.StatusInternalServerError {
			slog.Error(
				"Error while generating the Jwt Token",
				slog.Any("Error", err),
			)
			return "", "", "", "", err, errJson, errorCode
		}
	}

	slog.Debug(
		"JWT Token generated successfully",
		slog.String("Email", email),
	)
	return jwttoken, userID, userEmail, role, nil, nil, 0
}
