package models

import "time"

type Rating struct {
	ID        int       `json:"id"`
	RecipeID  int       `json:"recipe_id"`
	UserID    int       `json:"user_id"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RatingStats struct {
	RecipeID        int         `json:"recipe_id"`
	AverageRating   float64     `json:"average_rating"`
	TotalRatings    int         `json:"total_ratings"`
	RatingBreakdown map[int]int `json:"rating_breakdown"`
}

type Comment struct {
	ID        int       `json:"id"`
	RecipeID  int       `json:"recipe_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRatingRequest struct {
	Rating int `json:"rating"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}
