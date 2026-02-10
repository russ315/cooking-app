package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"cooking-app/internal/logger"
	"cooking-app/internal/middleware"
	"cooking-app/internal/models"
	"cooking-app/internal/recipe"
	"cooking-app/internal/repository"

	"github.com/gorilla/mux"
)

type RecipeHandler struct {
	repo            *repository.RecipeRepository
	search          *recipe.SearchService
	enhancedSearch  *recipe.EnhancedSearchService
	logger          *logger.ActivityLogger
}

func NewRecipeHandler(repo *repository.RecipeRepository, search *recipe.SearchService, enhancedSearch *recipe.EnhancedSearchService, log *logger.ActivityLogger) *RecipeHandler {
	return &RecipeHandler{
		repo:           repo,
		search:         search,
		enhancedSearch: enhancedSearch,
		logger:         log,
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

	userID := middleware.MustGetUserID(r)
	created := h.repo.Create(&req, userID)
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

	userID := middleware.MustGetUserID(r)
	updated, err := h.repo.Update(id, &req, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecipeForbidden) {
			http.Error(w, "Recipe can only be changed by its creator", http.StatusForbidden)
			return
		}
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

	userID := middleware.MustGetUserID(r)
	if err := h.repo.Delete(id, userID); err != nil {
		if errors.Is(err, repository.ErrRecipeForbidden) {
			http.Error(w, "Recipe can only be deleted by its creator", http.StatusForbidden)
			return
		}
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

// AdvancedIngredientSearch - POST /api/recipes/search/advanced
func (h *RecipeHandler) AdvancedIngredientSearch(w http.ResponseWriter, r *http.Request) {
	var req recipe.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default values
	if req.MaxResults <= 0 {
		req.MaxResults = 20
	}

	response := h.enhancedSearch.ComprehensiveSearch(req)
	h.logger.Log("advanced_search", 0)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetIngredientSubstitutes - GET /api/ingredients/{name}/substitutes
func (h *RecipeHandler) GetIngredientSubstitutes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ingredientName := vars["name"]
	if ingredientName == "" {
		http.Error(w, "Ingredient name is required", http.StatusBadRequest)
		return
	}

	substitutes := h.enhancedSearch.GetIngredientSubstitutes(ingredientName)
	h.logger.Log("ingredient_substitutes_viewed", 0)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"substitutes": substitutes})
}

// GetIngredientSynonyms - GET /api/ingredients/{name}/synonyms
func (h *RecipeHandler) GetIngredientSynonyms(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ingredientName := vars["name"]
	if ingredientName == "" {
		http.Error(w, "Ingredient name is required", http.StatusBadRequest)
		return
	}

	synonyms := h.enhancedSearch.GetIngredientSynonyms(ingredientName)
	h.logger.Log("ingredient_synonyms_viewed", 0)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"synonyms": synonyms})
}

// AddIngredientSynonym - POST /api/ingredients/synonyms
func (h *RecipeHandler) AddIngredientSynonym(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Canonical string `json:"canonical"`
		Synonym   string `json:"synonym"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Canonical == "" || req.Synonym == "" {
		http.Error(w, "Both canonical and synonym are required", http.StatusBadRequest)
		return
	}

	h.enhancedSearch.AddIngredientSynonym(req.Canonical, req.Synonym)
	h.logger.Log("ingredient_synonym_added", 0)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Synonym added successfully"})
}

// AddIngredientSubstitute - POST /api/ingredients/substitutes
func (h *RecipeHandler) AddIngredientSubstitute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Ingredient string `json:"ingredient"`
		Substitute string `json:"substitute"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Ingredient == "" || req.Substitute == "" {
		http.Error(w, "Both ingredient and substitute are required", http.StatusBadRequest)
		return
	}

	h.enhancedSearch.AddIngredientSubstitute(req.Ingredient, req.Substitute)
	h.logger.Log("ingredient_substitute_added", 0)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Substitute added successfully"})
}
