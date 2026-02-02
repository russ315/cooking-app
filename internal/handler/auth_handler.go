package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"cooking-app/internal/auth"
	"cooking-app/internal/models"
	"cooking-app/internal/repository"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	userRepo    *repository.UserRepository
	authService *auth.Service
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(userRepo *repository.UserRepository, authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		authService: authService,
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrWeakPassword) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.userRepo.CreateWithPassword(
		req.Username,
		req.Email,
		hashedPassword,
		req.FirstName,
		req.LastName,
	)
	if err != nil {
		if errors.Is(err, repository.ErrUsernameExists) {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, repository.ErrEmailExists) {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(models.AuthResponse{
		Token: token,
		User:  user,
	})
	if err != nil {
		return
	}
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Try to find user by username or email
	var user *models.User
	var err error

	// Check if it's an email (contains @)
	if strings.Contains(req.Username, "@") {
		user, err = h.userRepo.GetByEmail(req.Username)
	} else {
		user, err = h.userRepo.GetByUsername(req.Username)
	}

	if err != nil {
		if err == repository.ErrUserNotFound {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to find user", http.StatusInternalServerError)
		return
	}

	// Compare password
	if err := h.authService.ComparePassword(user.Password, req.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(models.AuthResponse{
		Token: token,
		User:  user,
	})
	if err != nil {
		return
	}
}
