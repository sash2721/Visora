package handlers

import (
	"Backend/errors"
	"Backend/services"
	"encoding/json"
	"log/slog"
	"net/http"
)

// Request structure
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response structure
type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"userID"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type AuthHandler struct {
	Service *services.AuthService
}

func (s *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Only POST method allowed"}`))
		return
	}

	var loginReq LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)

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

	// calling the login service to get the auth token
	token, userID, userEmail, role, serviceErr, errJson, errCode := s.Service.Login(loginReq.Email, loginReq.Password)

	if serviceErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errCode)
		w.Write(errJson)

		slog.Debug("Login Failed!", slog.String("Email", loginReq.Email))
		return
	}

	response := AuthResponse{Token: token, UserID: userID, Email: userEmail, Role: role}
	responseJson, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)

	slog.Info("Login Successful!", slog.String("Email", loginReq.Email))
}

func (s *AuthHandler) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Only POST method allowed"}`))
		return
	}

	var signupReq LoginRequest
	err := json.NewDecoder(r.Body).Decode(&signupReq)

	if err != nil {
		slog.Debug(
			"Incorrect signup request object sent!",
			slog.String("Email", signupReq.Email),
		)

		errJson, badRequestError := errors.NewBadRequestError("Incorrect signup request object sent!", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(badRequestError.Code)
		w.Write(errJson)

		return
	}

	token, userID, userEmail, role, serviceErr, errJson, errCode := s.Service.Signup(signupReq.Email, signupReq.Password)

	if serviceErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errCode)
		w.Write(errJson)

		slog.Debug("Signup Failed!", slog.String("Email", signupReq.Email))
		return
	}

	response := AuthResponse{Token: token, UserID: userID, Email: userEmail, Role: role}
	responseJson, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)

	slog.Info("Signup Successful!", slog.String("Email", signupReq.Email))
}
