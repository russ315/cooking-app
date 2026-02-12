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
	fmt.Println("+ Ratings & Comments System")
	fmt.Println("===========================================")
	fmt.Println()

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

	if err := db.Migrate(database); err != nil {
		log.Fatal("Database migrate failed:", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
		fmt.Println("⚠ Using default JWT_SECRET (set JWT_SECRET env var in production)")
	}

	userRepo := repository.NewUserRepository(database)
	recipeRepo := repository.NewRecipeRepository(database)
	ratingRepo := repository.NewRatingRepository(database)
	activityLogger := logger.NewActivityLogger()
	searchService := recipe.NewSearchService(recipeRepo)
	enhancedSearchService := recipe.NewEnhancedSearchService(recipeRepo)
	authService := auth.NewService(jwtSecret)

	authHandler := handler.NewAuthHandler(userRepo, authService)
	userHandler := handler.NewUserHandler(userRepo, activityLogger)
	recipeHandler := handler.NewRecipeHandler(recipeRepo, searchService, enhancedSearchService, activityLogger)
	ratingHandler := handler.NewRatingHandler(ratingRepo, activityLogger)

	authMiddleware := middleware.NewAuthMiddleware(authService)
	corsMiddleware := middleware.NewCORSMiddleware([]string{"*"}) // Allow all origins (change in production)

	router := mux.NewRouter()

	router.Use(corsMiddleware.Handler)

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	router.HandleFunc("/api/profiles", userHandler.GetAllProfiles).Methods("GET")
	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.GetProfile).Methods("GET")

	protectedProfile := router.PathPrefix("/api/profile").Subrouter()
	protectedProfile.Use(authMiddleware.Authenticate)
	protectedProfile.HandleFunc("", userHandler.CreateProfile).Methods("POST")
	protectedProfile.HandleFunc("/{id:[0-9]+}", userHandler.UpdateProfile).Methods("PUT")
	protectedProfile.HandleFunc("/{id:[0-9]+}", userHandler.DeleteProfile).Methods("DELETE")

	router.HandleFunc("/api/recipes", recipeHandler.ListRecipes).Methods("GET")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.GetRecipe).Methods("GET")
	router.HandleFunc("/api/ingredients", recipeHandler.ListIngredients).Methods("GET")

	router.HandleFunc("/api/recipes/search/advanced", recipeHandler.AdvancedIngredientSearch).Methods("POST")
	router.HandleFunc("/api/ingredients/{name}/substitutes", recipeHandler.GetIngredientSubstitutes).Methods("GET")
	router.HandleFunc("/api/ingredients/{name}/synonyms", recipeHandler.GetIngredientSynonyms).Methods("GET")

	router.HandleFunc("/api/recipes/{id:[0-9]+}/ratings", ratingHandler.GetRatingsByRecipe).Methods("GET")
	router.HandleFunc("/api/recipes/{id:[0-9]+}/rating-stats", ratingHandler.GetRatingStats).Methods("GET")
	router.HandleFunc("/api/recipes/{id:[0-9]+}/comments", ratingHandler.GetCommentsByRecipe).Methods("GET")

	protectedRecipes := router.PathPrefix("/api/recipes").Subrouter()
	protectedRecipes.Use(authMiddleware.Authenticate)
	protectedRecipes.HandleFunc("", recipeHandler.CreateRecipe).Methods("POST")
	protectedRecipes.HandleFunc("/{id:[0-9]+}", recipeHandler.UpdateRecipe).Methods("PUT")
	protectedRecipes.HandleFunc("/{id:[0-9]+}", recipeHandler.DeleteRecipe).Methods("DELETE")

	protectedRecipes.HandleFunc("/{id:[0-9]+}/ratings", ratingHandler.CreateOrUpdateRating).Methods("POST")
	protectedRecipes.HandleFunc("/{id:[0-9]+}/my-rating", ratingHandler.GetUserRatingForRecipe).Methods("GET")
	protectedRecipes.HandleFunc("/{id:[0-9]+}/comments", ratingHandler.CreateComment).Methods("POST")

	protectedIngredients := router.PathPrefix("/api/ingredients").Subrouter()
	protectedIngredients.Use(authMiddleware.Authenticate)
	protectedIngredients.HandleFunc("/synonyms", recipeHandler.AddIngredientSynonym).Methods("POST")
	protectedIngredients.HandleFunc("/substitutes", recipeHandler.AddIngredientSubstitute).Methods("POST")

	protectedComments := router.PathPrefix("/api/comments").Subrouter()
	protectedComments.Use(authMiddleware.Authenticate)
	protectedComments.HandleFunc("/{id:[0-9]+}", ratingHandler.UpdateComment).Methods("PUT")
	protectedComments.HandleFunc("/{id:[0-9]+}", ratingHandler.DeleteComment).Methods("DELETE")

	frontendFS := http.FileServer(http.Dir("./internal/frontend"))
	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			r.URL.Path = "/cooking-app-frontend.html"
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
	fmt.Println("    GET    /api/recipes/{id}/ratings           - Get all ratings for recipe")
	fmt.Println("    GET    /api/recipes/{id}/rating-stats      - Get rating statistics")
	fmt.Println("    GET    /api/recipes/{id}/comments          - Get all comments for recipe")
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
	fmt.Println("    POST   /api/recipes/{id}/ratings    - Create/update rating")
	fmt.Println("    GET    /api/recipes/{id}/my-rating  - Get your rating for recipe")
	fmt.Println("    POST   /api/recipes/{id}/comments   - Create comment")
	fmt.Println("    PUT    /api/comments/{id}           - Update comment")
	fmt.Println("    DELETE /api/comments/{id}           - Delete comment")
	fmt.Println()
	fmt.Println("  🌐 CORS enabled for all origins")
	fmt.Println("  🧠 Enhanced ingredient matching with fuzzy search, synonyms, and substitutes")
	fmt.Println("  ⭐ Recipe Rating & Comments System")
	fmt.Println()

	port := "8080"
	fmt.Printf("🚀 Server starting on http://localhost:%s\n", port)
	fmt.Println("   Main App: Visit http://localhost:" + port + "/")
	fmt.Println()

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
