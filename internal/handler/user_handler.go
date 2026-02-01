package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cooking-app/internal/logger"
	"cooking-app/internal/models"
	"cooking-app/internal/repository"

	"github.com/gorilla/mux"
)

// UserHandler обрабатывает HTTP запросы для User Profile API
type UserHandler struct {
	repo   *repository.UserRepository
	logger *logger.ActivityLogger
}

// NewUserHandler создает новый handler
func NewUserHandler(repo *repository.UserRepository, log *logger.ActivityLogger) *UserHandler {
	return &UserHandler{
		repo:   repo,
		logger: log,
	}
}

// GetProfile - GET /api/profile/{id}
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	h.logger.Log("profile_viewed", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetAllProfiles - GET /api/profiles
func (h *UserHandler) GetAllProfiles(w http.ResponseWriter, r *http.Request) {
	users := h.repo.GetAll()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// CreateProfile - POST /api/profile
func (h *UserHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created := h.repo.Create(&user)

	h.logger.Log("profile_created", created.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// UpdateProfile - PUT /api/profile/{id}
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.repo.Update(id, &req)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	h.logger.Log("profile_updated", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DeleteProfile - DELETE /api/profile/{id}
func (h *UserHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	h.logger.Log("profile_deleted", id)

	w.WriteHeader(http.StatusNoContent)
}
