package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

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
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL DEFAULT '',
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
			user_id INT REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS recipe_ingredients (
			recipe_id INT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
			ingredient_id INT NOT NULL REFERENCES ingredients(id),
			quantity TEXT NOT NULL,
			PRIMARY KEY (recipe_id, ingredient_id)
		)`,
		`CREATE TABLE IF NOT EXISTS ratings (
			id SERIAL PRIMARY KEY,
			recipe_id INT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
			user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(recipe_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			recipe_id INT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
			user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}

	if err := addPasswordColumnIfMissing(db); err != nil {
		return err
	}

	if err := addUniqueConstraintsIfMissing(db); err != nil {
		return err
	}

	if err := addRecipeUserIDIfMissing(db); err != nil {
		return err
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_recipes_name ON recipes(name)`,
		`CREATE INDEX IF NOT EXISTS idx_ratings_recipe ON ratings(recipe_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ratings_user ON ratings(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_recipe ON comments(recipe_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id)`,
	}
	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			log.Printf("Warning: could not create index: %v", err)
		}
	}

	log.Println("✓ Database tables created/verified (including ratings and comments)")
	return nil
}

func addPasswordColumnIfMissing(db *sql.DB) error {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'users' AND column_name = 'password'
		)
	`).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN password TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("add password column: %w", err)
		}
		log.Println("✓ Password column added to users table")
	}

	return nil
}

func addRecipeUserIDIfMissing(db *sql.DB) error {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'recipes' AND column_name = 'user_id'
		)
	`).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		if _, err := db.Exec(`ALTER TABLE recipes ADD COLUMN user_id INT REFERENCES users(id) ON DELETE SET NULL`); err != nil {
			return fmt.Errorf("add recipes.user_id column: %w", err)
		}
		log.Println("✓ recipes.user_id column added")
	}
	return nil
}

func addUniqueConstraintsIfMissing(db *sql.DB) error {
	var usernameUnique bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_constraint 
			WHERE conname = 'users_username_key'
		)
	`).Scan(&usernameUnique)
	if err == nil && !usernameUnique {
		if _, err := db.Exec(`ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username)`); err != nil {
			log.Printf("Warning: could not add unique constraint on username: %v", err)
		}
	}

	var emailUnique bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_constraint 
			WHERE conname = 'users_email_key'
		)
	`).Scan(&emailUnique)
	if err == nil && !emailUnique {
		if _, err := db.Exec(`ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email)`); err != nil {
			log.Printf("Warning: could not add unique constraint on email: %v", err)
		}
	}

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

	// Seed one user if no users exist (with hashed password)
	// Password: "test123456" - bcrypt hash
	var userCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err == nil && userCount == 0 {
		// This is bcrypt hash for "test123456"
		hashedPassword := "$2a$10$rQCd7e8K3k8K3k8K3k8K3eO.dZvZvZvZvZvZvZvZvZvZvZvZvZvZu"
		if _, err := db.Exec(`INSERT INTO users (username, email, password, first_name, last_name, bio, created_at)
			VALUES ('john_doe', 'john@example.com', $1, 'John', 'Doe', 'Test user', NOW())`, hashedPassword); err != nil {
			log.Println("Seed user:", err)
		} else {
			log.Println("✓ Sample user seeded (username: john_doe, password: test123456)")
		}
	}

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
