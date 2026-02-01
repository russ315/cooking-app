package main

import (
	"fmt"
	"log"
	"net/http"

	"cooking-app/internal/handler"
	"cooking-app/internal/logger"
	"cooking-app/internal/repository"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("Cooking App - User Profile Management API")
	fmt.Println("Assignment 4 - Milestone 2")
	fmt.Println("Author: Abilmansur")
	fmt.Println("===========================================")
	fmt.Println()

	userRepo := repository.NewUserRepository()
	fmt.Println("✓ In-memory user repository initialized")

	activityLogger := logger.NewActivityLogger()
	fmt.Println("✓ Activity logger with goroutine started")

	userHandler := handler.NewUserHandler(userRepo, activityLogger)

	router := mux.NewRouter()

	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.GetProfile).Methods("GET")
	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.UpdateProfile).Methods("PUT")
	router.HandleFunc("/api/profile/{id:[0-9]+}", userHandler.DeleteProfile).Methods("DELETE")
	router.HandleFunc("/api/profiles", userHandler.GetAllProfiles).Methods("GET")
	router.HandleFunc("/api/profile", userHandler.CreateProfile).Methods("POST")

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	fmt.Println()
	fmt.Println(" Available endpoints:")
	fmt.Println("  GET    /health               - Health check")
	fmt.Println("  GET    /api/profiles         - Get all profiles")
	fmt.Println("  GET    /api/profile/{id}     - Get profile by ID")
	fmt.Println("  POST   /api/profile          - Create new profile")
	fmt.Println("  PUT    /api/profile/{id}     - Update profile")
	fmt.Println("  DELETE /api/profile/{id}     - Delete profile")
	fmt.Println()

	port := "8080"
	fmt.Printf("🚀 Server starting on http://localhost:%s\n", port)
	fmt.Println()

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
