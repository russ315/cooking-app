package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func runReLink() {
	// Database connection
	connURL := os.Getenv("DATABASE_URL")
	if connURL == "" {
		connURL = "postgresql://postgres:Ruslan2006%40@localhost:5432/cooking?sslmode=disable"
		fmt.Println("Using default DATABASE_URL (set DATABASE_URL to override)")
	}
	
	db, err := sql.Open("pgx", connURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Re-link ingredients to recipes
	recipeIngredients := map[int][]string{
		1: {"egg", "butter", "salt"},           // Scrambled Eggs
		2: {"flour", "milk", "egg", "butter"},  // Pancakes
		3: {"chicken", "tomato", "onion", "garlic"}, // Tomato Chicken
		4: {"egg", "butter", "salt"},           // Scrambled Eggs (duplicate)
	}

	fmt.Println("Re-linking ingredients to recipes...")
	
	for recipeID, ingredients := range recipeIngredients {
		for _, ingredientName := range ingredients {
			// Get ingredient ID
			var ingredientID int
			err := db.QueryRow("SELECT id FROM ingredients WHERE name = $1", ingredientName).Scan(&ingredientID)
			if err != nil {
				fmt.Printf("Ingredient not found: %s\n", ingredientName)
				continue
			}

			// Insert recipe_ingredient link
			_, err = db.Exec("INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity) VALUES ($1, $2, $3)", 
				recipeID, ingredientID, "to taste")
			if err != nil {
				fmt.Printf("Error linking ingredient %s to recipe %d: %v\n", ingredientName, recipeID, err)
			} else {
				fmt.Printf("Linked %s to recipe %d\n", ingredientName, recipeID)
			}
		}
	}

	fmt.Println("Ingredients re-linked successfully!")
}
