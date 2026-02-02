package repository

import (
	"database/sql"
	"errors"
	"strings"
	"sync"
	"time"

	"recipe-backend/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

// UserRepository interface defines the contract for user data operations
type UserRepository interface {
	Create(user *models.User) error
	FindByID(id int) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindAll() ([]*models.User, error)
	Update(user *models.User) error
	Delete(id int) error
}

// InMemoryUserRepository implements UserRepository with thread-safe in-memory storage
type InMemoryUserRepository struct {
	mu         sync.RWMutex
	users      map[int]*models.User
	emailIndex map[string]int
	usernameIndex map[string]int
	nextID     int
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:      make(map[int]*models.User),
		emailIndex: make(map[string]int),
		usernameIndex: make(map[string]int),
		nextID:     1,
	}
}

// Create adds a new user to the repository
func (r *InMemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	normalizedEmail := strings.ToLower(user.Email)
	if _, exists := r.emailIndex[normalizedEmail]; exists {
		return ErrEmailAlreadyExists
	}

	// Check if username already exists
	normalizedUsername := strings.ToLower(user.Username)
	if _, exists := r.usernameIndex[normalizedUsername]; exists {
		return ErrUsernameAlreadyExists
	}

	// Assign ID and timestamps
	user.ID = r.nextID
	r.nextID++
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Store user
	r.users[user.ID] = user
	r.emailIndex[normalizedEmail] = user.ID
	r.usernameIndex[normalizedUsername] = user.ID

	return nil
}

// FindByID retrieves a user by their ID
func (r *InMemoryUserRepository) FindByID(id int) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// FindByEmail retrieves a user by their email
func (r *InMemoryUserRepository) FindByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedEmail := strings.ToLower(email)
	userID, exists := r.emailIndex[normalizedEmail]
	if !exists {
		return nil, ErrUserNotFound
	}

	return r.users[userID], nil
}

// FindByUsername retrieves a user by their username
func (r *InMemoryUserRepository) FindByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalizedUsername := strings.ToLower(username)
	userID, exists := r.usernameIndex[normalizedUsername]
	if !exists {
		return nil, ErrUserNotFound
	}

	return r.users[userID], nil
}

// FindAll retrieves all users
func (r *InMemoryUserRepository) FindAll() ([]*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []*models.User
	for rows.Next() {
		var u models.User
		var firstName, lastName, bio sql.NullString
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &firstName, &lastName, &bio, &u.CreatedAt); err != nil {
			continue
		}
		u.FirstName = firstName.String
		u.LastName = lastName.String
		u.Bio = bio.String
		users = append(users, &u)
	}

	return users, nil
}

// Update updates an existing user
func (r *InMemoryUserRepository) Update(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existingUser, exists := r.users[user.ID]
	if !exists {
		return ErrUserNotFound
	}

	// Update timestamp
	user.UpdatedAt = time.Now()
	user.CreatedAt = existingUser.CreatedAt // Preserve original creation time

	// Update user
	r.users[user.ID] = user

	return nil
}

// Delete removes a user from the repository
func (r *InMemoryUserRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return ErrUserNotFound
	}

	// Remove from indexes
	normalizedEmail := strings.ToLower(user.Email)
	normalizedUsername := strings.ToLower(user.Username)
	delete(r.emailIndex, normalizedEmail)
	delete(r.usernameIndex, normalizedUsername)
	delete(r.users, id)

	return nil
}

// RecipeRepository interface (placeholder for future implementation)
type RecipeRepository interface {
	Create(recipe *models.Recipe) error
	FindByID(id int) (*models.Recipe, error)
	FindAll() ([]*models.Recipe, error)
	Search(query string) ([]*models.Recipe, error)
}

// IngredientRepository interface (placeholder for future implementation)
type IngredientRepository interface {
	Create(ingredient *models.Ingredient) error
	FindByID(id int) (*models.Ingredient, error)
	FindAll() ([]*models.Ingredient, error)
}
