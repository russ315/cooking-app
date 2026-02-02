// Alternative entry: use root main.go and run "go run ." from project root.
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"recipe-backend/config"
	"recipe-backend/internal/handlers"
	"recipe-backend/internal/middleware"
	"recipe-backend/internal/repository"
	"recipe-backend/internal/service"
	"recipe-backend/pkg/jwt"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize JWT manager
	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TokenDuration)

	// Initialize repositories
	userRepo := repository.NewInMemoryUserRepository()

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtManager)

	"cooking-app/internal/db"
	"cooking-app/internal/handler"
	"cooking-app/internal/logger"
	"cooking-app/internal/recipe"
	"cooking-app/internal/repository"

	"github.com/gorilla/mux"

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

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Setup router
	mux := http.NewServeMux()

	// Public routes (no authentication required)
	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.Register(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.Login(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/auth/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authHandler.ValidateToken(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Protected routes (authentication required)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/auth/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authHandler.GetProfile(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "healthy",
			"service": "recipe-backend",
			"timestamp": "` + time.Now().Format(time.RFC3339) + `",
			"version": "1.0.0"
		}`))
	})

	// Root endpoint with API documentation
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"message": "Recipe Management API",
			"version": "1.0.0",
			"endpoints": {
				"auth": {
					"register": "POST /api/auth/register",
					"login": "POST /api/auth/login",
					"validate": "GET /api/auth/validate",
					"profile": "GET /api/auth/profile (requires auth)"
				},
				"health": "GET /health"
			},
			"documentation": "See README.md for detailed API documentation"
		}`))
	})

	// Combine public and protected routes
	mux.Handle("/api/auth/profile", middleware.AuthMiddleware(jwtManager)(protectedMux))

	// Apply global middleware
	handler := middleware.RecoveryMiddleware(
		middleware.Logger(
			middleware.CORS(
				middleware.RateLimiter(mux),
			),
		),
	)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on http://%s:%s", cfg.Server.Host, cfg.Server.Port)
		log.Printf("📝 Environment: %s", cfg.Server.Env)
		log.Printf("💚 Health check: http://%s:%s/health", cfg.Server.Host, cfg.Server.Port)
		log.Printf("🔐 Auth endpoints:")
		log.Printf("   - POST http://%s:%s/api/auth/register", cfg.Server.Host, cfg.Server.Port)
		log.Printf("   - POST http://%s:%s/api/auth/login", cfg.Server.Host, cfg.Server.Port)
		log.Printf("   - GET  http://%s:%s/api/auth/validate", cfg.Server.Host, cfg.Server.Port)
		log.Printf("   - GET  http://%s:%s/api/auth/profile (protected)", cfg.Server.Host, cfg.Server.Port)
		log.Println("Press Ctrl+C to stop the server")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop

	log.Println("\n🛑 Shutting down server gracefully...")

	// Shutdown auth service background processes
	authService.Shutdown()

	log.Println("✅ Server stopped")
}
