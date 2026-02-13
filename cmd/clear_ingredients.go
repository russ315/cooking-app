package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func run() {
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

	// Clear all ingredients (need to clear recipe_ingredients first due to foreign key)
	fmt.Println("Clearing recipe_ingredients table...")
	_, err = db.Exec("DELETE FROM recipe_ingredients")
	if err != nil {
		log.Fatal("Failed to clear recipe_ingredients:", err)
	}

	fmt.Println("Clearing ingredients table...")
	_, err = db.Exec("DELETE FROM ingredients")
	if err != nil {
		log.Fatal("Failed to clear ingredients:", err)
	}

	// Reset sequence
	_, err = db.Exec("ALTER SEQUENCE ingredients_id_seq RESTART WITH 1")
	if err != nil {
		log.Fatal("Failed to reset sequence:", err)
	}

	fmt.Println("Ingredients cleared successfully!")
}
