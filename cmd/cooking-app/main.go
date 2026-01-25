package main

import (
	"log"
	"net/http"
)

func main() {
	// TODO: Initialize database connection
	// TODO: Set up routes and middleware
	// TODO: Start HTTP server

	log.Println("Cooking App Server starting...")
	
	// Placeholder server
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
