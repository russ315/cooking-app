// Package main is the entry point for Assignment 4.
// Run from project root: go run .
package main

import (
	"fmt"
	"log"
	"net/http"

	"cooking-app/internal/handler"
	"cooking-app/internal/logger"
	"cooking-app/internal/recipe"
	"cooking-app/internal/repository"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("Cooking App - Assignment 4 Milestone 2")
	fmt.Println("Recipe Search Logic + User Profile API")
	fmt.Println("===========================================")
	fmt.Println()

	userRepo := repository.NewUserRepository()
	recipeRepo := repository.NewRecipeRepository()
	activityLogger := logger.NewActivityLogger()
	searchService := recipe.NewSearchService(recipeRepo)

	userHandler := handler.NewUserHandler(userRepo, activityLogger)
	recipeHandler := handler.NewRecipeHandler(recipeRepo, searchService, activityLogger)

	router := mux.NewRouter()

	// Health
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// User profile API
	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.GetProfile).Methods("GET")
	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.UpdateProfile).Methods("PUT")
	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.DeleteProfile).Methods("DELETE")
	router.HandleFunc("/api/profiles", userHandler.GetAllProfiles).Methods("GET")
	router.HandleFunc("/api/profile", userHandler.CreateProfile).Methods("POST")

	// Recipe Search API (Assignment 4 - Recipe Search Logic)
	router.HandleFunc("/api/recipes", recipeHandler.ListRecipes).Methods("GET")
	router.HandleFunc("/api/recipes", recipeHandler.CreateRecipe).Methods("POST")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.GetRecipe).Methods("GET")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.UpdateRecipe).Methods("PUT")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.DeleteRecipe).Methods("DELETE")
	router.HandleFunc("/api/ingredients", recipeHandler.ListIngredients).Methods("GET")

	fmt.Println(" Endpoints:")
	fmt.Println("  GET    /health                 - Health check")
	fmt.Println("  GET    /api/profiles           - Get all user profiles")
	fmt.Println("  GET    /api/profile/{id}       - Get profile by ID")
	fmt.Println("  POST   /api/profile            - Create profile")
	fmt.Println("  PUT    /api/profile/{id}       - Update profile")
	fmt.Println("  DELETE /api/profile/{id}       - Delete profile")
	fmt.Println("  GET    /api/recipes            - List recipes (optional ?search=... or ?ingredients=egg,flour)")
	fmt.Println("  GET    /api/recipes/{id}       - Get recipe by ID")
	fmt.Println("  POST   /api/recipes           - Create recipe")
	fmt.Println("  PUT    /api/recipes/{id}      - Update recipe")
	fmt.Println("  DELETE /api/recipes/{id}      - Delete recipe")
	fmt.Println("  GET    /api/ingredients       - List ingredients")
	fmt.Println()

	port := "8080"
	fmt.Printf(" Server starting on http://localhost:%s\n", port)
	fmt.Println()

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
