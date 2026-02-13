package repository

import (
	"database/sql"
	"fmt"

	"cooking-app/internal/models"
)

// IngredientRepository manages ingredients in the database.
type IngredientRepository struct {
	db *sql.DB
}

// NewIngredientRepository creates a new ingredient repository.
func NewIngredientRepository(db *sql.DB) *IngredientRepository {
	return &IngredientRepository{db: db}
}

// CreateIngredient creates a new ingredient if it doesn't exist.
func (r *IngredientRepository) CreateIngredient(name string) (*models.Ingredient, error) {
	// Check if ingredient already exists
	var existingID int
	err := r.db.QueryRow("SELECT id FROM ingredients WHERE LOWER(name) = LOWER($1)", name).Scan(&existingID)
	if err == nil {
		// Ingredient exists
		return &models.Ingredient{ID: existingID, Name: name}, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Insert new ingredient
	var id int
	err = r.db.QueryRow("INSERT INTO ingredients (name) VALUES ($1) RETURNING id", name).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &models.Ingredient{ID: id, Name: name}, nil
}

// InitializeIngredients adds common ingredients to the database.
func (r *IngredientRepository) InitializeIngredients() error {
	ingredients := map[string][]string{
		"egg":       {"eggs"},
		"flour":     {},
		"sugar":     {},
		"butter":    {},
		"milk":      {},
		"onion":     {},
		"garlic":    {},
		"tomato":    {"tomatoes"},
		"potato":    {"potatoes"},
		"carrot":    {"carrots"},
		"chicken":   {},
		"beef":      {},
		"rice":      {},
		"pasta":     {},
		"cheese":    {},
		"olive oil": {},
		"salt":      {},
		"pepper":    {},
	}

	for canonical, synonyms := range ingredients {
		// Create main ingredient
		_, err := r.CreateIngredient(canonical)
		if err != nil {
			fmt.Printf("Error creating ingredient %s: %v\n", canonical, err)
		}

		// Create synonyms
		for _, synonym := range synonyms {
			_, err := r.CreateIngredient(synonym)
			if err != nil {
				fmt.Printf("Error creating ingredient %s: %v\n", synonym, err)
			}
		}
	}

	return nil
}

// GetIngredientByName finds an ingredient by name (case-insensitive).
func (r *IngredientRepository) GetIngredientByName(name string) (*models.Ingredient, error) {
	var ing models.Ingredient
	err := r.db.QueryRow("SELECT id, name FROM ingredients WHERE LOWER(name) = LOWER($1)", name).Scan(&ing.ID, &ing.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ingredient not found: %s", name)
		}
		return nil, err
	}
	return &ing, nil
}

// GetAllIngredients returns all ingredients.
func (r *IngredientRepository) GetAllIngredients() ([]models.Ingredient, error) {
	rows, err := r.db.Query("SELECT id, name FROM ingredients ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		var ing models.Ingredient
		if err := rows.Scan(&ing.ID, &ing.Name); err != nil {
			continue
		}
		ingredients = append(ingredients, ing)
	}
	return ingredients, nil
}
