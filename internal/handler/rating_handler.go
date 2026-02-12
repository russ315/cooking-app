package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"cooking-app/internal/logger"
	"cooking-app/internal/middleware"
	"cooking-app/internal/models"
	"cooking-app/internal/repository"

	"github.com/gorilla/mux"
)

type RatingHandler struct {
	repo   *repository.RatingRepository
	logger *logger.ActivityLogger
}

func NewRatingHandler(repo *repository.RatingRepository, log *logger.ActivityLogger) *RatingHandler {
	return &RatingHandler{
		repo:   repo,
		logger: log,
	}
}

func (h *RatingHandler) CreateOrUpdateRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req models.CreateRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	userID := middleware.MustGetUserID(r)
	rating, err := h.repo.CreateOrUpdateRating(recipeID, userID, req.Rating)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Log("rating_created_or_updated", recipeID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rating)
}

func (h *RatingHandler) GetRatingsByRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	ratings, err := h.repo.GetRatingsByRecipe(recipeID)
	if err != nil {
		http.Error(w, "Failed to fetch ratings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ratings)
}

func (h *RatingHandler) GetRatingStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	stats, err := h.repo.GetRatingStats(recipeID)
	if err != nil {
		http.Error(w, "Failed to fetch rating stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *RatingHandler) GetUserRatingForRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	userID := middleware.MustGetUserID(r)
	rating, err := h.repo.GetUserRatingForRecipe(recipeID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRatingNotFound) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"rating": nil})
			return
		}
		http.Error(w, "Failed to fetch rating", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rating)
}

func (h *RatingHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Comment content cannot be empty", http.StatusBadRequest)
		return
	}

	userID := middleware.MustGetUserID(r)
	comment, err := h.repo.CreateComment(recipeID, userID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Log("comment_created", recipeID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (h *RatingHandler) GetCommentsByRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	comments, err := h.repo.GetCommentsByRecipe(recipeID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (h *RatingHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Comment content cannot be empty", http.StatusBadRequest)
		return
	}

	userID := middleware.MustGetUserID(r)
	comment, err := h.repo.UpdateComment(commentID, userID, req.Content)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrCommentForbidden) {
			http.Error(w, "You can only edit your own comments", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Log("comment_updated", commentID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

func (h *RatingHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	userID := middleware.MustGetUserID(r)
	err = h.repo.DeleteComment(commentID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrCommentForbidden) {
			http.Error(w, "You can only delete your own comments", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Log("comment_deleted", commentID)

	w.WriteHeader(http.StatusNoContent)
}
