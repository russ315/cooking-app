package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open connects to PostgreSQL and returns *sql.DB (thread-safe pool).
// connURL example: "postgres://user:pass@localhost:5432/cooking?sslmode=disable"
func Open(connURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

// Migrate creates tables and seeds initial data if empty.
func Migrate(db *sql.DB) error {
	if err := createTables(db); err != nil {
		return err
	}
	return seedIfEmpty(db)
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL,
			email TEXT NOT NULL,
			first_name TEXT,
			last_name TEXT,
			bio TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ingredients (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS recipes (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			instructions TEXT,
			prep_time_min INT NOT NULL DEFAULT 0,
			cook_time_min INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS recipe_ingredients (
			recipe_id INT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
			ingredient_id INT NOT NULL REFERENCES ingredients(id),
			quantity TEXT NOT NULL,
			PRIMARY KEY (recipe_id, ingredient_id)
		)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}
	log.Println("✓ Database tables created/verified")
	return nil
}

func seedIfEmpty(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM ingredients").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	ingNames := []string{"Eggs", "Flour", "Milk", "Butter", "Sugar", "Salt", "Chicken", "Tomato", "Onion", "Garlic"}
	for _, name := range ingNames {
		if _, err := db.Exec("INSERT INTO ingredients (name) VALUES ($1)", name); err != nil {
			return fmt.Errorf("seed ingredient: %w", err)
		}
	}
	log.Println("✓ Ingredients seeded")

	// Seed one user if no users exist
	var userCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err == nil && userCount == 0 {
		if _, err := db.Exec(`INSERT INTO users (username, email, first_name, last_name, bio, created_at)
			VALUES ('john_doe', 'john@example.com', 'John', 'Doe', 'Test user', NOW())`); err != nil {
			log.Println("Seed user:", err)
		} else {
			log.Println("✓ Sample user seeded")
		}
	}

	// Seed sample recipes with recipe_ingredients
	type recIng struct {
		ingID int
		qty   string
	}
	recipes := []struct {
		name, desc, instructions string
		prep, cook               int
		ingredients              []recIng
	}{
		{"Scrambled Eggs", "Simple fluffy scrambled eggs", "Beat eggs, add salt. Cook in butter on low heat, stir gently.", 2, 5,
			[]recIng{{1, "3"}, {6, "pinch"}, {4, "1 tbsp"}}},
		{"Pancakes", "Classic breakfast pancakes", "Mix flour, milk, eggs. Cook on griddle until bubbles form, flip.", 5, 10,
			[]recIng{{1, "2"}, {2, "1 cup"}, {3, "1 cup"}, {5, "2 tbsp"}, {4, "2 tbsp"}}},
		{"Tomato Chicken", "Chicken with tomato and garlic", "Brown chicken, add onion and garlic, add tomato. Simmer 20 min.", 10, 25,
			[]recIng{{7, "500g"}, {8, "2"}, {9, "1"}, {10, "2 cloves"}}},
	}
	for _, r := range recipes {
		var recipeID int
		if err := db.QueryRow(`INSERT INTO recipes (name, description, instructions, prep_time_min, cook_time_min)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`, r.name, r.desc, r.instructions, r.prep, r.cook).Scan(&recipeID); err != nil {
			return fmt.Errorf("seed recipe: %w", err)
		}
		for _, ri := range r.ingredients {
			if _, err := db.Exec(`INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity)
				VALUES ($1, $2, $3)`, recipeID, ri.ingID, ri.qty); err != nil {
				return fmt.Errorf("seed recipe_ingredient: %w", err)
			}
		}
	}
	log.Println("✓ Sample recipes seeded")
	return nil
}
