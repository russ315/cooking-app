package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"cooking-app/internal/repository"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func runInit() {
	// Database connection - use same logic as main.go
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

	// Create ingredient repository
	ingRepo := repository.NewIngredientRepository(db)

	// Initialize all ingredients
	fmt.Println("Initializing ingredients...")
	err = ingRepo.InitializeIngredients()
	if err != nil {
		log.Fatal("Failed to initialize ingredients:", err)
	}

	fmt.Println("Ingredients initialized successfully!")
}
