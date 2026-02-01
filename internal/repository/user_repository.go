package repository

import (
	"errors"
	"sync"
	"time"

	"cooking-app/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository хранит пользователей в памяти (без БД)
type UserRepository struct {
	mu     sync.RWMutex // Mutex для thread-safe доступа (Assignment 4)
	users  map[int]*models.User
	nextID int
}

// NewUserRepository создает новый репозиторий
func NewUserRepository() *UserRepository {
	repo := &UserRepository{
		users:  make(map[int]*models.User),
		nextID: 1,
	}

	// Добавляем тестовые данные
	repo.users[1] = &models.User{
		ID:        1,
		Username:  "john_doe",
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Bio:       "Test user",
		CreatedAt: time.Now(),
	}
	repo.nextID = 2

	return repo
}

// GetByID получает пользователя по ID (thread-safe с RLock)
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetAll возвращает всех пользователей
func (r *UserRepository) GetAll() []*models.User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users
}

// Create создает нового пользователя (thread-safe с Lock)
func (r *UserRepository) Create(user *models.User) *models.User {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.ID = r.nextID
	user.CreatedAt = time.Now()
	r.users[user.ID] = user
	r.nextID++

	return user
}

// Update обновляет пользователя (thread-safe с Lock)
func (r *UserRepository) Update(id int, req *models.UpdateUserRequest) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Bio = req.Bio

	return user, nil
}

// Delete удаляет пользователя (thread-safe с Lock)
func (r *UserRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return ErrUserNotFound
	}

	delete(r.users, id)
	return nil
}
