package models

import "time"

// User represents a user profile with authentication.
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never send in JSON
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Bio       string    `json:"bio,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateUserRequest for updating user profile.
type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}

// RegisterRequest for user registration.
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// LoginRequest for user login.
type LoginRequest struct {
	Username string `json:"username"` // can be username or email
	Password string `json:"password"`
}

// AuthResponse returned after successful login/register.
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
