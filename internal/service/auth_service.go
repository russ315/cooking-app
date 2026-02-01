package service

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"recipe-backend/internal/models"
	"recipe-backend/internal/repository"
	"recipe-backend/pkg/jwt"
	"recipe-backend/pkg/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo       repository.UserRepository
	jwtManager     *jwt.JWTManager
	loginAttempts  map[string]*LoginAttemptTracker
	attemptsMutex  sync.RWMutex
	cleanupTicker  *time.Ticker
	stopCleanup    chan bool
}

// LoginAttemptTracker tracks failed login attempts for rate limiting
type LoginAttemptTracker struct {
	Count     int
	LastAttempt time.Time
	LockedUntil *time.Time
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, jwtManager *jwt.JWTManager) *AuthService {
	service := &AuthService{
		userRepo:      userRepo,
		jwtManager:    jwtManager,
		loginAttempts: make(map[string]*LoginAttemptTracker),
		cleanupTicker: time.NewTicker(5 * time.Minute),
		stopCleanup:   make(chan bool),
	}

	// Start background goroutine for cleanup
	go service.cleanupLoginAttempts()

	return service
}

// Register handles user registration
func (s *AuthService) Register(req models.RegisterRequest) (*models.AuthResponse, error) {
	// Validate input
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, err
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	if err := utils.ValidateUsername(req.Username); err != nil {
		return nil, err
	}

	// Normalize email
	normalizedEmail := utils.NormalizeEmail(req.Email)

	// Check if user already exists
	if _, err := s.userRepo.FindByEmail(normalizedEmail); err == nil {
		return nil, repository.ErrEmailAlreadyExists
	}

	// Check if username already exists
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return nil, repository.ErrUsernameAlreadyExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     utils.SanitizeInput(req.Username),
		Email:        normalizedEmail,
		PasswordHash: hashedPassword,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Log registration event (background processing)
	go s.logRegistrationEvent(user.ID, user.Email)

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Login handles user authentication
func (s *AuthService) Login(req models.LoginRequest) (*models.AuthResponse, error) {
	normalizedEmail := utils.NormalizeEmail(req.Email)

	// Check if account is locked
	if s.isAccountLocked(normalizedEmail) {
		return nil, errors.New("account temporarily locked due to multiple failed login attempts")
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(normalizedEmail)
	if err != nil {
		s.recordFailedLogin(normalizedEmail)
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		s.recordFailedLogin(normalizedEmail)
		return nil, ErrInvalidCredentials
	}

	// Clear failed login attempts on successful login
	s.clearFailedAttempts(normalizedEmail)

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Log login event (background processing)
	go s.logLoginEvent(user.ID, user.Email, true)

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(id int) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

// recordFailedLogin tracks failed login attempts
func (s *AuthService) recordFailedLogin(email string) {
	s.attemptsMutex.Lock()
	defer s.attemptsMutex.Unlock()

	tracker, exists := s.loginAttempts[email]
	if !exists {
		tracker = &LoginAttemptTracker{
			Count:       0,
			LastAttempt: time.Now(),
		}
		s.loginAttempts[email] = tracker
	}

	tracker.Count++
	tracker.LastAttempt = time.Now()

	// Lock account after 5 failed attempts for 15 minutes
	if tracker.Count >= 5 {
		lockUntil := time.Now().Add(15 * time.Minute)
		tracker.LockedUntil = &lockUntil
		log.Printf("Account locked for email: %s until %v", email, lockUntil)
	}
}

// clearFailedAttempts clears failed login attempts
func (s *AuthService) clearFailedAttempts(email string) {
	s.attemptsMutex.Lock()
	defer s.attemptsMutex.Unlock()

	delete(s.loginAttempts, email)
}

// isAccountLocked checks if an account is temporarily locked
func (s *AuthService) isAccountLocked(email string) bool {
	s.attemptsMutex.RLock()
	defer s.attemptsMutex.RUnlock()

	tracker, exists := s.loginAttempts[email]
	if !exists {
		return false
	}

	if tracker.LockedUntil != nil && time.Now().Before(*tracker.LockedUntil) {
		return true
	}

	return false
}

// cleanupLoginAttempts runs as a background goroutine to clean up old login attempts
func (s *AuthService) cleanupLoginAttempts() {
	for {
		select {
		case <-s.cleanupTicker.C:
			s.attemptsMutex.Lock()
			now := time.Now()
			for email, tracker := range s.loginAttempts {
				// Remove entries older than 1 hour
				if now.Sub(tracker.LastAttempt) > 1*time.Hour {
					delete(s.loginAttempts, email)
				}
				// Clear lock if time has passed
				if tracker.LockedUntil != nil && now.After(*tracker.LockedUntil) {
					tracker.LockedUntil = nil
					tracker.Count = 0
				}
			}
			s.attemptsMutex.Unlock()
			log.Println("Cleaned up old login attempts")
		case <-s.stopCleanup:
			s.cleanupTicker.Stop()
			return
		}
	}
}

// logRegistrationEvent logs registration events (simulates async processing)
func (s *AuthService) logRegistrationEvent(userID int, email string) {
	// Simulate some processing delay
	time.Sleep(100 * time.Millisecond)
	log.Printf("[REGISTRATION EVENT] User ID: %d, Email: %s, Timestamp: %s",
		userID, email, time.Now().Format(time.RFC3339))
}

// logLoginEvent logs login events (simulates async processing)
func (s *AuthService) logLoginEvent(userID int, email string, success bool) {
	// Simulate some processing delay
	time.Sleep(100 * time.Millisecond)
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	log.Printf("[LOGIN EVENT] User ID: %d, Email: %s, Status: %s, Timestamp: %s",
		userID, email, status, time.Now().Format(time.RFC3339))
}

// Shutdown gracefully shuts down the service
func (s *AuthService) Shutdown() {
	s.stopCleanup <- true
}
