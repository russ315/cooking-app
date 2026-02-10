// Package main is the entry point for Assignment 4.
// Run from project root: go run .
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"cooking-app/internal/auth"
	"cooking-app/internal/db"
	"cooking-app/internal/handler"
	"cooking-app/internal/logger"
	"cooking-app/internal/middleware"
	"cooking-app/internal/recipe"
	"cooking-app/internal/repository"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("Cooking App - With Authentication + CORS")
	fmt.Println("Recipe Search + User Profiles + Auth (JWT)")
	fmt.Println("===========================================")
	fmt.Println()

	// Database connection
	connURL := os.Getenv("DATABASE_URL")
	if connURL == "" {
		connURL = "postgres://postgres:postgres@localhost:5432/cooking?sslmode=disable"
		fmt.Println("Using default DATABASE_URL (set DATABASE_URL to override)")
	}
	database, err := db.Open(connURL)
	if err != nil {
		fmt.Println()
		fmt.Println("PostgreSQL connection failed. Common causes:")
		fmt.Println("  - Wrong password: set DATABASE_URL with your postgres user and password")
		fmt.Println("  - Example (PowerShell): $env:DATABASE_URL=\"postgres://postgres:YOUR_PASSWORD@localhost:5432/cooking?sslmode=disable\"")
		fmt.Println("  - Create database first: createdb cooking")
		fmt.Println()
		log.Fatal("Database connection failed:", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.Migrate(database); err != nil {
		log.Fatal("Database migrate failed:", err)
	}

	// Initialize services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
		fmt.Println("⚠ Using default JWT_SECRET (set JWT_SECRET env var in production)")
	}

	userRepo := repository.NewUserRepository(database)
	recipeRepo := repository.NewRecipeRepository(database)
	activityLogger := logger.NewActivityLogger()
	searchService := recipe.NewSearchService(recipeRepo)
	enhancedSearchService := recipe.NewEnhancedSearchService(recipeRepo)
	authService := auth.NewService(jwtSecret)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userRepo, authService)
	userHandler := handler.NewUserHandler(userRepo, activityLogger)
	recipeHandler := handler.NewRecipeHandler(recipeRepo, searchService, enhancedSearchService, activityLogger)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	corsMiddleware := middleware.NewCORSMiddleware([]string{"*"}) // Allow all origins (change in production)

	// Setup router
	router := mux.NewRouter()

	// Apply CORS middleware to all routes
	router.Use(corsMiddleware.Handler)

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Public auth routes
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	// Public profile routes (can view profiles without auth)
	router.HandleFunc("/api/profiles", userHandler.GetAllProfiles).Methods("GET")
	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.GetProfile).Methods("GET")

	// Protected profile routes (require authentication)
	protectedProfile := router.PathPrefix("/api/profile").Subrouter()
	protectedProfile.Use(authMiddleware.Authenticate)
	protectedProfile.HandleFunc("", userHandler.CreateProfile).Methods("POST")
	protectedProfile.HandleFunc("/{id:[0-9]+}", userHandler.UpdateProfile).Methods("PUT")
	protectedProfile.HandleFunc("/{id:[0-9]+}", userHandler.DeleteProfile).Methods("DELETE")

	// Recipe routes (public read, protected write)
	router.HandleFunc("/api/recipes", recipeHandler.ListRecipes).Methods("GET")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.GetRecipe).Methods("GET")
	router.HandleFunc("/api/ingredients", recipeHandler.ListIngredients).Methods("GET")

	// Enhanced ingredient matching routes (public)
	router.HandleFunc("/api/recipes/search/advanced", recipeHandler.AdvancedIngredientSearch).Methods("POST")
	router.HandleFunc("/api/ingredients/{name}/substitutes", recipeHandler.GetIngredientSubstitutes).Methods("GET")
	router.HandleFunc("/api/ingredients/{name}/synonyms", recipeHandler.GetIngredientSynonyms).Methods("GET")

	// Protected recipe routes
	protectedRecipes := router.PathPrefix("/api/recipes").Subrouter()
	protectedRecipes.Use(authMiddleware.Authenticate)
	protectedRecipes.HandleFunc("", recipeHandler.CreateRecipe).Methods("POST")
	protectedRecipes.HandleFunc("/{id:[0-9]+}", recipeHandler.UpdateRecipe).Methods("PUT")
	protectedRecipes.HandleFunc("/{id:[0-9]+}", recipeHandler.DeleteRecipe).Methods("DELETE")

	// Protected ingredient management routes
	protectedIngredients := router.PathPrefix("/api/ingredients").Subrouter()
	protectedIngredients.Use(authMiddleware.Authenticate)
	protectedIngredients.HandleFunc("/synonyms", recipeHandler.AddIngredientSynonym).Methods("POST")
	protectedIngredients.HandleFunc("/substitutes", recipeHandler.AddIngredientSubstitute).Methods("POST")

	// Frontend: serve React single-page app for all non-API routes
	frontendFS := http.FileServer(http.Dir("./internal/frontend"))
	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route to different pages based on path
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			r.URL.Path = "/cooking-app-frontend.html" // Main app with login/register
		}
		frontendFS.ServeHTTP(w, r)
	}))

	fmt.Println("📋 API Endpoints:")
	fmt.Println()
	fmt.Println("  PUBLIC:")
	fmt.Println("    GET    /health                      - Health check")
	fmt.Println("    POST   /api/auth/register           - Register new user")
	fmt.Println("    POST   /api/auth/login              - Login user")
	fmt.Println("    GET    /api/profiles                - Get all profiles")
	fmt.Println("    GET    /api/profile/{id}            - Get profile by ID")
	fmt.Println("    GET    /api/recipes                 - List recipes (search: ?search=... or ?ingredients=...)")
	fmt.Println("    GET    /api/recipes/{id}            - Get recipe by ID")
	fmt.Println("    GET    /api/ingredients             - List ingredients")
	fmt.Println("    POST   /api/recipes/search/advanced - Advanced ingredient matching")
	fmt.Println("    GET    /api/ingredients/{name}/substitutes - Get ingredient substitutes")
	fmt.Println("    GET    /api/ingredients/{name}/synonyms     - Get ingredient synonyms")
	fmt.Println()
	fmt.Println("  PROTECTED (require Authorization: Bearer <token>):")
	fmt.Println("    POST   /api/profile                 - Create profile")
	fmt.Println("    PUT    /api/profile/{id}            - Update profile")
	fmt.Println("    DELETE /api/profile/{id}            - Delete profile")
	fmt.Println("    POST   /api/recipes                 - Create recipe")
	fmt.Println("    PUT    /api/recipes/{id}            - Update recipe")
	fmt.Println("    DELETE /api/recipes/{id}            - Delete recipe")
	fmt.Println("    POST   /api/ingredients/synonyms    - Add ingredient synonym")
	fmt.Println("    POST   /api/ingredients/substitutes - Add ingredient substitute")
	fmt.Println()
	fmt.Println("  🌐 CORS enabled for all origins")
	fmt.Println("  🧠 Enhanced ingredient matching with fuzzy search, synonyms, and substitutes")
	fmt.Println()

	port := "8080"
	fmt.Printf("🚀 Server starting on http://localhost:%s\n", port)
	fmt.Println("   Main App: Visit http://localhost:" + port + "/")
	fmt.Println()

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
