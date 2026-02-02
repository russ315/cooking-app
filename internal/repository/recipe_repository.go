package repository

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"cooking-app/internal/models"
)

var (
	ErrRecipeNotFound = errors.New("recipe not found")
)

// RecipeRepository stores recipes and ingredients in PostgreSQL.
type RecipeRepository struct {
	db *sql.DB
}

// NewRecipeRepository creates a new repository backed by PostgreSQL.
func NewRecipeRepository(db *sql.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

// scanRecipe scans a recipe row and loads ingredients in a second query (or we could use a join and group).
func (r *RecipeRepository) scanRecipe(row *sql.Row) (*models.Recipe, error) {
	var rec models.Recipe
	var desc, instructions sql.NullString
	err := row.Scan(&rec.ID, &rec.Name, &desc, &instructions, &rec.PrepTimeMin, &rec.CookTimeMin, &rec.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecipeNotFound
		}
		return nil, err
	}
	rec.Description = desc.String
	rec.Instructions = instructions.String
	rec.Ingredients, _ = r.loadIngredients(rec.ID)
	return &rec, nil
}

func (r *RecipeRepository) loadIngredients(recipeID int) ([]models.RecipeIngredient, error) {
	rows, err := r.db.Query(`SELECT ri.recipe_id, ri.ingredient_id, ri.quantity, i.name
		FROM recipe_ingredients ri JOIN ingredients i ON i.id = ri.ingredient_id
		WHERE ri.recipe_id = $1 ORDER BY ri.ingredient_id`, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.RecipeIngredient
	for rows.Next() {
		var ri models.RecipeIngredient
		var name string
		if err := rows.Scan(&ri.RecipeID, &ri.IngredientID, &ri.Quantity, &name); err != nil {
			continue
		}
		ri.Ingredient = models.Ingredient{ID: ri.IngredientID, Name: name}
		list = append(list, ri)
	}
	return list, nil
}

// GetByID returns a recipe by ID with ingredients.
func (r *RecipeRepository) GetByID(id int) (*models.Recipe, error) {
	row := r.db.QueryRow(`SELECT id, name, description, instructions, prep_time_min, cook_time_min, created_at
		FROM recipes WHERE id = $1`, id)
	return r.scanRecipe(row)
}

// GetAll returns all recipes with ingredients.
func (r *RecipeRepository) GetAll() []*models.Recipe {
	rows, err := r.db.Query(`SELECT id, name, description, instructions, prep_time_min, cook_time_min, created_at FROM recipes ORDER BY id`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var list []*models.Recipe
	for rows.Next() {
		var rec models.Recipe
		var desc, instructions sql.NullString
		if err := rows.Scan(&rec.ID, &rec.Name, &desc, &instructions, &rec.PrepTimeMin, &rec.CookTimeMin, &rec.CreatedAt); err != nil {
			continue
		}
		rec.Description = desc.String
		rec.Instructions = instructions.String
		rec.Ingredients, _ = r.loadIngredients(rec.ID)
		list = append(list, &rec)
	}
	return list
}

// Create inserts a new recipe and its ingredients.
func (r *RecipeRepository) Create(req *models.CreateRecipeRequest) *models.Recipe {
	var id int
	var createdAt time.Time
	err := r.db.QueryRow(`INSERT INTO recipes (name, description, instructions, prep_time_min, cook_time_min)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		req.Name, req.Description, req.Instructions, req.PrepTimeMin, req.CookTimeMin).Scan(&id, &createdAt)
	if err != nil {
		return nil
	}

	for _, ri := range req.Ingredients {
		r.db.Exec(`INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity) VALUES ($1, $2, $3)`,
			id, ri.IngredientID, ri.Quantity)
	}

	created, _ := r.GetByID(id)
	return created
}

// Update updates recipe and replaces its ingredients.
func (r *RecipeRepository) Update(id int, req *models.UpdateRecipeRequest) (*models.Recipe, error) {
	res, err := r.db.Exec(`UPDATE recipes SET name = $1, description = $2, instructions = $3, prep_time_min = $4, cook_time_min = $5 WHERE id = $6`,
		req.Name, req.Description, req.Instructions, req.PrepTimeMin, req.CookTimeMin, id)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrRecipeNotFound
	}

	if _, err := r.db.Exec("DELETE FROM recipe_ingredients WHERE recipe_id = $1", id); err != nil {
		return nil, err
	}
	for _, ri := range req.Ingredients {
		r.db.Exec(`INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity) VALUES ($1, $2, $3)`,
			id, ri.IngredientID, ri.Quantity)
	}
	return r.GetByID(id)
}

// Delete removes a recipe (cascade deletes recipe_ingredients).
func (r *RecipeRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM recipes WHERE id = $1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrRecipeNotFound
	}
	return nil
}

// SearchByName returns recipes whose name or description contains the query (case-insensitive).
func (r *RecipeRepository) SearchByName(query string) []*models.Recipe {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return r.GetAll()
	}
	pattern := "%" + query + "%"
	rows, err := r.db.Query(`SELECT id, name, description, instructions, prep_time_min, cook_time_min, created_at
		FROM recipes WHERE LOWER(name) LIKE $1 OR LOWER(COALESCE(description,'')) LIKE $2 ORDER BY id`, pattern, pattern)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var list []*models.Recipe
	for rows.Next() {
		var rec models.Recipe
		var desc, instructions sql.NullString
		if err := rows.Scan(&rec.ID, &rec.Name, &desc, &instructions, &rec.PrepTimeMin, &rec.CookTimeMin, &rec.CreatedAt); err != nil {
			continue
		}
		rec.Description = desc.String
		rec.Instructions = instructions.String
		rec.Ingredients, _ = r.loadIngredients(rec.ID)
		list = append(list, &rec)
	}
	return list
}

// SearchByIngredients returns recipes that contain ALL of the given ingredient names.
func (r *RecipeRepository) SearchByIngredients(ingredientNames []string) []*models.Recipe {
	if len(ingredientNames) == 0 {
		return r.GetAll()
	}
	want := make(map[string]bool)
	for _, n := range ingredientNames {
		n = strings.TrimSpace(strings.ToLower(n))
		if n != "" {
			want[n] = true
		}
	}
	if len(want) == 0 {
		return r.GetAll()
	}

	// Recipe IDs that have ALL of the wanted ingredients (HAVING COUNT = len(want)).
	args := make([]interface{}, 0, len(want)+1)
	inParts := make([]string, 0, len(want))
	pos := 1
	for name := range want {
		args = append(args, name)
		inParts = append(inParts, "$"+strconv.Itoa(pos))
		pos++
	}
	args = append(args, len(want))
	inPart := "LOWER(i.name) IN (" + strings.Join(inParts, ",") + ")"
	q := `SELECT ri.recipe_id FROM recipe_ingredients ri JOIN ingredients i ON i.id = ri.ingredient_id WHERE ` + inPart + ` GROUP BY ri.recipe_id HAVING COUNT(DISTINCT LOWER(i.name)) = $` + strconv.Itoa(pos)
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}

	// Load full recipes
	var list []*models.Recipe
	for _, id := range ids {
		rec, _ := r.GetByID(id)
		if rec != nil {
			list = append(list, rec)
		}
	}
	return list
}

// ListIngredients returns all ingredients.
func (r *RecipeRepository) ListIngredients() []*models.Ingredient {
	rows, err := r.db.Query("SELECT id, name FROM ingredients ORDER BY id")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var list []*models.Ingredient
	for rows.Next() {
		var ing models.Ingredient
		if err := rows.Scan(&ing.ID, &ing.Name); err != nil {
			continue
		}
		list = append(list, &ing)
	}
	return list
}
