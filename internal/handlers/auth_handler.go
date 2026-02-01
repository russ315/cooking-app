package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"recipe-backend/internal/models"
	"recipe-backend/internal/repository"
	"recipe-backend/internal/service"
	"recipe-backend/pkg/utils"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call service
	authResp, err := h.authService.Register(req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Send response
	h.respondJSON(w, http.StatusCreated, authResp)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call service
	authResp, err := h.authService.Login(req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Send response
	h.respondJSON(w, http.StatusOK, authResp)
}

// GetProfile returns the authenticated user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get user from service
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Send response
	h.respondJSON(w, http.StatusOK, user)
}

// ValidateToken validates a JWT token
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.respondError(w, http.StatusUnauthorized, "Missing authorization header")
		return
	}

	// Extract token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.respondError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}

	token := parts[1]

	// Validate token
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Send response
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid": true,
		"user": map[string]interface{}{
			"id":       claims.UserID,
			"email":    claims.Email,
			"username": claims.Username,
		},
	})
}

// handleServiceError maps service errors to HTTP responses
func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error) {
	log.Printf("Service error: %v", err)

	switch {
	case errors.Is(err, repository.ErrEmailAlreadyExists):
		h.respondError(w, http.StatusConflict, "Email already registered")
	case errors.Is(err, repository.ErrUsernameAlreadyExists):
		h.respondError(w, http.StatusConflict, "Username already taken")
	case errors.Is(err, service.ErrInvalidCredentials):
		h.respondError(w, http.StatusUnauthorized, "Invalid email or password")
	case errors.Is(err, utils.ErrInvalidEmail):
		h.respondError(w, http.StatusBadRequest, utils.FormatValidationError(err))
	case errors.Is(err, utils.ErrWeakPassword):
		h.respondError(w, http.StatusBadRequest, utils.FormatValidationError(err))
	case errors.Is(err, utils.ErrInvalidUsername):
		h.respondError(w, http.StatusBadRequest, utils.FormatValidationError(err))
	case strings.Contains(err.Error(), "account temporarily locked"):
		h.respondError(w, http.StatusTooManyRequests, err.Error())
	default:
		h.respondError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// respondJSON sends a JSON response
func (h *AuthHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// respondError sends an error response
func (h *AuthHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, models.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
