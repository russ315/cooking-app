package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Recipe represents a recipe
type Recipe struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Instructions string    `json:"instructions"`
	CookTime     int       `json:"cook_time"` // in minutes
	Servings     int       `json:"servings"`
	CreatedBy    int       `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Ingredient represents an ingredient
type Ingredient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"` // kg, g, ml, l, pieces, etc.
}

// RecipeIngredient represents the many-to-many relationship between recipes and ingredients
type RecipeIngredient struct {
	RecipeID     int     `json:"recipe_id"`
	IngredientID int     `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
}

// UserFavorite represents user's favorite recipes
type UserFavorite struct {
	UserID   int       `json:"user_id"`
	RecipeID int       `json:"recipe_id"`
	SavedAt  time.Time `json:"saved_at"`
}

// RecipeRating represents user ratings for recipes
type RecipeRating struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	RecipeID  int       `json:"recipe_id"`
	Rating    int       `json:"rating"` // 1-5 stars
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// UserInventory represents user's ingredient inventory
type UserInventory struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	IngredientID int       `json:"ingredient_id"`
	Quantity     float64   `json:"quantity"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DTO structs for requests/responses

// RegisterRequest for user registration
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse returned after successful authentication
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ErrorResponse for error messages
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse for generic success messages
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
