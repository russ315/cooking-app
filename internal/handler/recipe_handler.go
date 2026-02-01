package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"cooking-app/internal/logger"
	"cooking-app/internal/models"
	"cooking-app/internal/recipe"
	"cooking-app/internal/repository"

	"github.com/gorilla/mux"
)

type RecipeHandler struct {
	repo   *repository.RecipeRepository
	search *recipe.SearchService
	logger *logger.ActivityLogger
}

func NewRecipeHandler(repo *repository.RecipeRepository, search *recipe.SearchService, log *logger.ActivityLogger) *RecipeHandler {
	return &RecipeHandler{
		repo:   repo,
		search: search,
		logger: log,
	}
}

// ListRecipes - GET /api/recipes (optional query: search=..., ingredients=...)
func (h *RecipeHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("search")
	ingredientsParam := r.URL.Query().Get("ingredients")

	var recipes []*models.Recipe
	if ingredientsParam != "" {
		names := strings.Split(ingredientsParam, ",")
		for i := range names {
			names[i] = strings.TrimSpace(names[i])
		}
		recipes = h.search.SearchByIngredients(names)
	} else if searchQuery != "" {
		recipes = h.search.SearchByName(searchQuery)
	} else {
		recipes = h.repo.GetAll()
	}

	h.logger.Log("recipes_listed", 0)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipes)
}

// GetRecipe - GET /api/recipes/{id}
func (h *RecipeHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	recipe, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	h.logger.Log("recipe_viewed", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipe)
}

// CreateRecipe - POST /api/recipes
func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	created := h.repo.Create(&req)
	h.search.NotifyRecipeChange(created.ID)
	h.logger.Log("recipe_created", created.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// UpdateRecipe - PUT /api/recipes/{id}
func (h *RecipeHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.repo.Update(id, &req)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	h.search.NotifyRecipeChange(id)
	h.logger.Log("recipe_updated", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DeleteRecipe - DELETE /api/recipes/{id}
func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	h.logger.Log("recipe_deleted", id)
	w.WriteHeader(http.StatusNoContent)
}

// ListIngredients - GET /api/ingredients
func (h *RecipeHandler) ListIngredients(w http.ResponseWriter, r *http.Request) {
	list := h.repo.ListIngredients()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
