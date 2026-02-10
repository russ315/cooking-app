package models

import "time"

type Recipe struct {
	ID           int               `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Instructions string            `json:"instructions"`
	PrepTimeMin  int               `json:"prep_time_min"`
	CookTimeMin  int               `json:"cook_time_min"`
	Ingredients  []RecipeIngredient `json:"ingredients"`
	UserID       *int               `json:"user_id,omitempty"` // creator; nil for legacy recipes
	CreatedAt    time.Time         `json:"created_at"`
}

type RecipeIngredient struct {
	RecipeID     int        `json:"recipe_id"`
	IngredientID int        `json:"ingredient_id"`
	Ingredient   Ingredient `json:"ingredient,omitempty"`
	Quantity     string     `json:"quantity"` // e.g. "2 cups", "100g"
}

type Ingredient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateRecipeRequest struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Instructions string            `json:"instructions"`
	PrepTimeMin  int               `json:"prep_time_min"`
	CookTimeMin  int               `json:"cook_time_min"`
	Ingredients  []RecipeIngredient `json:"ingredients"`
}

type UpdateRecipeRequest struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Instructions string            `json:"instructions"`
	PrepTimeMin  int               `json:"prep_time_min"`
	CookTimeMin  int               `json:"cook_time_min"`
	Ingredients  []RecipeIngredient `json:"ingredients"`
}
