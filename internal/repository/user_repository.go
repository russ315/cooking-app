package repository

import (
	"database/sql"
	"errors"
	"time"

	"cooking-app/internal/models"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUsernameExists = errors.New("username already exists")
	ErrEmailExists    = errors.New("email already exists")
)

// UserRepository stores users in PostgreSQL (thread-safe via connection pool).
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new repository backed by PostgreSQL.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID returns a user by ID.
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	row := r.db.QueryRow(`SELECT id, username, email, password, first_name, last_name, bio, created_at
		FROM users WHERE id = $1`, id)
	var u models.User
	var firstName, lastName, bio sql.NullString
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &firstName, &lastName, &bio, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.FirstName = firstName.String
	u.LastName = lastName.String
	u.Bio = bio.String
	return &u, nil
}

// GetByUsername returns a user by username.
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	row := r.db.QueryRow(`SELECT id, username, email, password, first_name, last_name, bio, created_at
		FROM users WHERE username = $1`, username)
	var u models.User
	var firstName, lastName, bio sql.NullString
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &firstName, &lastName, &bio, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.FirstName = firstName.String
	u.LastName = lastName.String
	u.Bio = bio.String
	return &u, nil
}

// GetByEmail returns a user by email.
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow(`SELECT id, username, email, password, first_name, last_name, bio, created_at
		FROM users WHERE email = $1`, email)
	var u models.User
	var firstName, lastName, bio sql.NullString
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &firstName, &lastName, &bio, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.FirstName = firstName.String
	u.LastName = lastName.String
	u.Bio = bio.String
	return &u, nil
}

// GetAll returns all users.
func (r *UserRepository) GetAll() []*models.User {
	rows, err := r.db.Query(`SELECT id, username, email, password, first_name, last_name, bio, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		var firstName, lastName, bio sql.NullString
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &firstName, &lastName, &bio, &u.CreatedAt); err != nil {
			continue
		}
		u.FirstName = firstName.String
		u.LastName = lastName.String
		u.Bio = bio.String
		users = append(users, &u)
	}
	return users
}

// Create inserts a new user (without password - for old API compatibility).
func (r *UserRepository) Create(user *models.User) *models.User {
	var id int
	var createdAt time.Time
	err := r.db.QueryRow(`INSERT INTO users (username, email, password, first_name, last_name, bio)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`,
		user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.Bio).Scan(&id, &createdAt)
	if err != nil {
		return nil
	}
	user.ID = id
	user.CreatedAt = createdAt
	return user
}

// CreateWithPassword inserts a new user with hashed password.
func (r *UserRepository) CreateWithPassword(username, email, hashedPassword, firstName, lastName string) (*models.User, error) {
	// Check if username exists
	existing, _ := r.GetByUsername(username)
	if existing != nil {
		return nil, ErrUsernameExists
	}

	// Check if email exists
	existing, _ = r.GetByEmail(email)
	if existing != nil {
		return nil, ErrEmailExists
	}

	var id int
	var createdAt time.Time
	err := r.db.QueryRow(`INSERT INTO users (username, email, password, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		username, email, hashedPassword, firstName, lastName).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        id,
		Username:  username,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		CreatedAt: createdAt,
	}, nil
}

// Update updates first_name, last_name, bio by ID.
func (r *UserRepository) Update(id int, req *models.UpdateUserRequest) (*models.User, error) {
	res, err := r.db.Exec(`UPDATE users SET first_name = $1, last_name = $2, bio = $3 WHERE id = $4`,
		req.FirstName, req.LastName, req.Bio, id)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrUserNotFound
	}
	return r.GetByID(id)
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}
