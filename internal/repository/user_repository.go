package repository

import (
	"database/sql"
	"errors"
	"time"

	"cooking-app/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
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
	row := r.db.QueryRow(`SELECT id, username, email, first_name, last_name, bio, created_at
		FROM users WHERE id = $1`, id)
	var u models.User
	var firstName, lastName, bio sql.NullString
	err := row.Scan(&u.ID, &u.Username, &u.Email, &firstName, &lastName, &bio, &u.CreatedAt)
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
	rows, err := r.db.Query(`SELECT id, username, email, first_name, last_name, bio, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()

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
	return users
}

// Create inserts a new user and returns it with ID and CreatedAt set.
func (r *UserRepository) Create(user *models.User) *models.User {
	var id int
	var createdAt time.Time
	err := r.db.QueryRow(`INSERT INTO users (username, email, first_name, last_name, bio)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		user.Username, user.Email, user.FirstName, user.LastName, user.Bio).Scan(&id, &createdAt)
	if err != nil {
		return nil
	}
	user.ID = id
	user.CreatedAt = createdAt
	return user
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
