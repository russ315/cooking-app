package repository

import (
	"database/sql"
	"errors"
	"time"

	"cooking-app/internal/models"
)

var (
	ErrRatingNotFound   = errors.New("rating not found")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCommentForbidden = errors.New("comment can only be modified by its author")
)

type RatingRepository struct {
	db *sql.DB
}

func NewRatingRepository(db *sql.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) CreateOrUpdateRating(recipeID, userID, rating int) (*models.Rating, error) {
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	var id int
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(`
		UPDATE ratings 
		SET rating = $1, updated_at = NOW() 
		WHERE recipe_id = $2 AND user_id = $3 
		RETURNING id, created_at, updated_at`,
		rating, recipeID, userID).Scan(&id, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		err = r.db.QueryRow(`
			INSERT INTO ratings (recipe_id, user_id, rating, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			RETURNING id, created_at, updated_at`,
			recipeID, userID, rating).Scan(&id, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &models.Rating{
		ID:        id,
		RecipeID:  recipeID,
		UserID:    userID,
		Rating:    rating,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *RatingRepository) GetRatingsByRecipe(recipeID int) ([]*models.Rating, error) {
	rows, err := r.db.Query(`
		SELECT id, recipe_id, user_id, rating, created_at, updated_at
		FROM ratings
		WHERE recipe_id = $1
		ORDER BY created_at DESC`, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []*models.Rating
	for rows.Next() {
		var rating models.Rating
		if err := rows.Scan(&rating.ID, &rating.RecipeID, &rating.UserID,
			&rating.Rating, &rating.CreatedAt, &rating.UpdatedAt); err != nil {
			continue
		}
		ratings = append(ratings, &rating)
	}

	return ratings, nil
}

func (r *RatingRepository) GetUserRatingForRecipe(recipeID, userID int) (*models.Rating, error) {
	var rating models.Rating
	err := r.db.QueryRow(`
		SELECT id, recipe_id, user_id, rating, created_at, updated_at
		FROM ratings
		WHERE recipe_id = $1 AND user_id = $2`, recipeID, userID).
		Scan(&rating.ID, &rating.RecipeID, &rating.UserID,
			&rating.Rating, &rating.CreatedAt, &rating.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrRatingNotFound
	}
	if err != nil {
		return nil, err
	}

	return &rating, nil
}

func (r *RatingRepository) GetRatingStats(recipeID int) (*models.RatingStats, error) {
	stats := &models.RatingStats{
		RecipeID:        recipeID,
		RatingBreakdown: make(map[int]int),
	}

	err := r.db.QueryRow(`
		SELECT COALESCE(AVG(rating), 0), COUNT(*)
		FROM ratings
		WHERE recipe_id = $1`, recipeID).
		Scan(&stats.AverageRating, &stats.TotalRatings)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`
		SELECT rating, COUNT(*)
		FROM ratings
		WHERE recipe_id = $1
		GROUP BY rating`, recipeID)
	if err != nil {
		return stats, nil
	}
	defer rows.Close()

	for rows.Next() {
		var rating, count int
		if err := rows.Scan(&rating, &count); err == nil {
			stats.RatingBreakdown[rating] = count
		}
	}

	return stats, nil
}

func (r *RatingRepository) CreateComment(recipeID, userID int, content string) (*models.Comment, error) {
	if content == "" {
		return nil, errors.New("comment content cannot be empty")
	}

	var id int
	var createdAt, updatedAt time.Time
	var username string

	err := r.db.QueryRow(`
		INSERT INTO comments (recipe_id, user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at`,
		recipeID, userID, content).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	r.db.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&username)

	return &models.Comment{
		ID:        id,
		RecipeID:  recipeID,
		UserID:    userID,
		Username:  username,
		Content:   content,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *RatingRepository) GetCommentsByRecipe(recipeID int) ([]*models.Comment, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.recipe_id, c.user_id, u.username, c.content, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.recipe_id = $1
		ORDER BY c.created_at DESC`, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.RecipeID, &comment.UserID,
			&comment.Username, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			continue
		}
		comments = append(comments, &comment)
	}

	return comments, nil
}

func (r *RatingRepository) GetCommentByID(id int) (*models.Comment, error) {
	var comment models.Comment
	var username string

	err := r.db.QueryRow(`
		SELECT c.id, c.recipe_id, c.user_id, u.username, c.content, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = $1`, id).
		Scan(&comment.ID, &comment.RecipeID, &comment.UserID,
			&username, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrCommentNotFound
	}
	if err != nil {
		return nil, err
	}

	comment.Username = username
	return &comment, nil
}

func (r *RatingRepository) UpdateComment(id, userID int, content string) (*models.Comment, error) {
	if content == "" {
		return nil, errors.New("comment content cannot be empty")
	}

	comment, err := r.GetCommentByID(id)
	if err != nil {
		return nil, err
	}

	if comment.UserID != userID {
		return nil, ErrCommentForbidden
	}

	_, err = r.db.Exec(`
		UPDATE comments 
		SET content = $1, updated_at = NOW()
		WHERE id = $2`, content, id)
	if err != nil {
		return nil, err
	}

	return r.GetCommentByID(id)
}

func (r *RatingRepository) DeleteComment(id, userID int) error {
	comment, err := r.GetCommentByID(id)
	if err != nil {
		return err
	}

	if comment.UserID != userID {
		return ErrCommentForbidden
	}

	res, err := r.db.Exec("DELETE FROM comments WHERE id = $1", id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrCommentNotFound
	}

	return nil
}
