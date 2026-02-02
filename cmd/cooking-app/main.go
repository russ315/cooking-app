// Alternative entry: use root main.go and run "go run ." from project root.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"cooking-app/internal/db"
	"cooking-app/internal/handler"
	"cooking-app/internal/logger"
	"cooking-app/internal/recipe"
	"cooking-app/internal/repository"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("Cooking App - Assignment 4 Milestone 2")
	fmt.Println("(Run from root: go run .)")
	fmt.Println("===========================================")
	fmt.Println()

	connURL := os.Getenv("DATABASE_URL")
	if connURL == "" {
		connURL = "postgres://postgres:postgres@localhost:5432/cooking?sslmode=disable"
	}
	database, err := db.Open(connURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer database.Close()
	if err := db.Migrate(database); err != nil {
		log.Fatal("Database migrate failed:", err)
	}

	userRepo := repository.NewUserRepository(database)
	recipeRepo := repository.NewRecipeRepository(database)
	activityLogger := logger.NewActivityLogger()
	searchService := recipe.NewSearchService(recipeRepo)

	userHandler := handler.NewUserHandler(userRepo, activityLogger)
	recipeHandler := handler.NewRecipeHandler(recipeRepo, searchService, activityLogger)

	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.GetProfile).Methods("GET")
	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.UpdateProfile).Methods("PUT")
	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.DeleteProfile).Methods("DELETE")
	router.HandleFunc("/api/profiles", userHandler.GetAllProfiles).Methods("GET")
	router.HandleFunc("/api/profile", userHandler.CreateProfile).Methods("POST")

	router.HandleFunc("/api/recipes", recipeHandler.ListRecipes).Methods("GET")
	router.HandleFunc("/api/recipes", recipeHandler.CreateRecipe).Methods("POST")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.GetRecipe).Methods("GET")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.UpdateRecipe).Methods("PUT")
	router.HandleFunc("/api/recipes/{id:[0-9]+}", recipeHandler.DeleteRecipe).Methods("DELETE")
	router.HandleFunc("/api/ingredients", recipeHandler.ListIngredients).Methods("GET")

	port := "8080"
	fmt.Printf(" Server starting on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
