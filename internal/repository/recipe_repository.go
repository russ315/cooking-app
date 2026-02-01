package repository

import (
	"errors"
	"strings"
	"sync"
	"time"

	"cooking-app/internal/models"
)

var (
	ErrRecipeNotFound = errors.New("recipe not found")
)

// RecipeRepository stores recipes and ingredients in memory (thread-safe, matches ERD)
type RecipeRepository struct {
	mu         sync.RWMutex
	recipes    map[int]*models.Recipe
	ingredients map[int]*models.Ingredient
	nextRecipeID int
	nextIngredientID int
}

// NewRecipeRepository creates a new repository with seed data
func NewRecipeRepository() *RecipeRepository {
	repo := &RecipeRepository{
		recipes:     make(map[int]*models.Recipe),
		ingredients: make(map[int]*models.Ingredient),
		nextRecipeID:    1,
		nextIngredientID: 1,
	}
	repo.seedData()
	return repo
}

func (r *RecipeRepository) seedData() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Seed ingredients
	ingNames := []string{"Eggs", "Flour", "Milk", "Butter", "Sugar", "Salt", "Chicken", "Tomato", "Onion", "Garlic"}
	for _, name := range ingNames {
		ing := &models.Ingredient{ID: r.nextIngredientID, Name: name}
		r.ingredients[r.nextIngredientID] = ing
		r.nextIngredientID++
	}

	// Seed recipes
	r.recipes[1] = &models.Recipe{
		ID:           1,
		Name:         "Scrambled Eggs",
		Description:  "Simple fluffy scrambled eggs",
		Instructions: "Beat eggs, add salt. Cook in butter on low heat, stir gently.",
		PrepTimeMin:  2,
		CookTimeMin:  5,
		Ingredients: []models.RecipeIngredient{
			{RecipeID: 1, IngredientID: 1, Quantity: "3", Ingredient: models.Ingredient{ID: 1, Name: "Eggs"}},
			{RecipeID: 1, IngredientID: 6, Quantity: "pinch", Ingredient: models.Ingredient{ID: 6, Name: "Salt"}},
			{RecipeID: 1, IngredientID: 4, Quantity: "1 tbsp", Ingredient: models.Ingredient{ID: 4, Name: "Butter"}},
		},
		CreatedAt: time.Now(),
	}
	r.recipes[2] = &models.Recipe{
		ID:           2,
		Name:         "Pancakes",
		Description:  "Classic breakfast pancakes",
		Instructions: "Mix flour, milk, eggs. Cook on griddle until bubbles form, flip.",
		PrepTimeMin:  5,
		CookTimeMin:  10,
		Ingredients: []models.RecipeIngredient{
			{RecipeID: 2, IngredientID: 1, Quantity: "2", Ingredient: models.Ingredient{ID: 1, Name: "Eggs"}},
			{RecipeID: 2, IngredientID: 2, Quantity: "1 cup", Ingredient: models.Ingredient{ID: 2, Name: "Flour"}},
			{RecipeID: 2, IngredientID: 3, Quantity: "1 cup", Ingredient: models.Ingredient{ID: 3, Name: "Milk"}},
			{RecipeID: 2, IngredientID: 5, Quantity: "2 tbsp", Ingredient: models.Ingredient{ID: 5, Name: "Sugar"}},
			{RecipeID: 2, IngredientID: 4, Quantity: "2 tbsp", Ingredient: models.Ingredient{ID: 4, Name: "Butter"}},
		},
		CreatedAt: time.Now(),
	}
	r.recipes[3] = &models.Recipe{
		ID:           3,
		Name:         "Tomato Chicken",
		Description:  "Chicken with tomato and garlic",
		Instructions: "Brown chicken, add onion and garlic, add tomato. Simmer 20 min.",
		PrepTimeMin:  10,
		CookTimeMin:  25,
		Ingredients: []models.RecipeIngredient{
			{RecipeID: 3, IngredientID: 7, Quantity: "500g", Ingredient: models.Ingredient{ID: 7, Name: "Chicken"}},
			{RecipeID: 3, IngredientID: 8, Quantity: "2", Ingredient: models.Ingredient{ID: 8, Name: "Tomato"}},
			{RecipeID: 3, IngredientID: 9, Quantity: "1", Ingredient: models.Ingredient{ID: 9, Name: "Onion"}},
			{RecipeID: 3, IngredientID: 10, Quantity: "2 cloves", Ingredient: models.Ingredient{ID: 10, Name: "Garlic"}},
		},
		CreatedAt: time.Now(),
	}
	r.nextRecipeID = 4
}

// GetByID returns a recipe by ID (thread-safe)
func (r *RecipeRepository) GetByID(id int) (*models.Recipe, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	recipe, exists := r.recipes[id]
	if !exists {
		return nil, ErrRecipeNotFound
	}
	return recipe, nil
}

// GetAll returns all recipes
func (r *RecipeRepository) GetAll() []*models.Recipe {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*models.Recipe, 0, len(r.recipes))
	for _, recipe := range r.recipes {
		list = append(list, recipe)
	}
	return list
}

// Create creates a new recipe (thread-safe)
func (r *RecipeRepository) Create(req *models.CreateRecipeRequest) *models.Recipe {
	r.mu.Lock()
	defer r.mu.Unlock()

	recipe := &models.Recipe{
		ID:           r.nextRecipeID,
		Name:         req.Name,
		Description:  req.Description,
		Instructions: req.Instructions,
		PrepTimeMin:  req.PrepTimeMin,
		CookTimeMin:  req.CookTimeMin,
		Ingredients:  req.Ingredients,
		CreatedAt:    time.Now(),
	}
	for i := range recipe.Ingredients {
		recipe.Ingredients[i].RecipeID = recipe.ID
		if ing, ok := r.ingredients[recipe.Ingredients[i].IngredientID]; ok {
			recipe.Ingredients[i].Ingredient = *ing
		}
	}
	r.recipes[recipe.ID] = recipe
	r.nextRecipeID++
	return recipe
}

// Update updates an existing recipe
func (r *RecipeRepository) Update(id int, req *models.UpdateRecipeRequest) (*models.Recipe, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	recipe, exists := r.recipes[id]
	if !exists {
		return nil, ErrRecipeNotFound
	}
	recipe.Name = req.Name
	recipe.Description = req.Description
	recipe.Instructions = req.Instructions
	recipe.PrepTimeMin = req.PrepTimeMin
	recipe.CookTimeMin = req.CookTimeMin
	recipe.Ingredients = req.Ingredients
	for i := range recipe.Ingredients {
		recipe.Ingredients[i].RecipeID = id
		if ing, ok := r.ingredients[recipe.Ingredients[i].IngredientID]; ok {
			recipe.Ingredients[i].Ingredient = *ing
		}
	}
	return recipe, nil
}

// Delete removes a recipe
func (r *RecipeRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.recipes[id]; !exists {
		return ErrRecipeNotFound
	}
	delete(r.recipes, id)
	return nil
}

// SearchByName returns recipes whose name contains the query (case-insensitive)
func (r *RecipeRepository) SearchByName(query string) []*models.Recipe {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return r.getAllLocked()
	}

	var result []*models.Recipe
	for _, recipe := range r.recipes {
		if strings.Contains(strings.ToLower(recipe.Name), query) ||
			strings.Contains(strings.ToLower(recipe.Description), query) {
			result = append(result, recipe)
		}
	}
	return result
}

// SearchByIngredients returns recipes that contain ALL of the given ingredient names (or any if empty)
func (r *RecipeRepository) SearchByIngredients(ingredientNames []string) []*models.Recipe {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(ingredientNames) == 0 {
		return r.getAllLocked()
	}

	// Normalize: lowercase, trim
	want := make(map[string]bool)
	for _, n := range ingredientNames {
		n = strings.TrimSpace(strings.ToLower(n))
		if n != "" {
			want[n] = true
		}
	}
	if len(want) == 0 {
		return r.getAllLocked()
	}

	var result []*models.Recipe
	for _, recipe := range r.recipes {
		have := make(map[string]bool)
		for _, ri := range recipe.Ingredients {
			have[strings.ToLower(ri.Ingredient.Name)] = true
		}
		allMatch := true
		for w := range want {
			if !have[w] {
				allMatch = false
				break
			}
		}
		if allMatch {
			result = append(result, recipe)
		}
	}
	return result
}

func (r *RecipeRepository) getAllLocked() []*models.Recipe {
	list := make([]*models.Recipe, 0, len(r.recipes))
	for _, recipe := range r.recipes {
		list = append(list, recipe)
	}
	return list
}

// ListIngredients returns all ingredients (for dropdowns etc.)
func (r *RecipeRepository) ListIngredients() []*models.Ingredient {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*models.Ingredient, 0, len(r.ingredients))
	for _, ing := range r.ingredients {
		list = append(list, ing)
	}
	return list
}
