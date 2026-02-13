package recipe

import (
	"strings"
	"sync"

	"cooking-app/internal/models"
)

// RecipeRepository interface for recipe operations
type RecipeRepository interface {
	GetAll() []*models.Recipe
	GetByID(id int) (*models.Recipe, error)
	Create(req *models.CreateRecipeRequest, userID int) *models.Recipe
	Update(id int, req *models.UpdateRecipeRequest, userID int) (*models.Recipe, error)
	Delete(id int, userID int) error
	SearchByName(query string) []*models.Recipe
	SearchByIngredients(names []string) []*models.Recipe
	ListIngredients() []*models.Ingredient
}

// EnhancedSearchService encapsulates advanced recipe search logic with ingredient matching
type EnhancedSearchService struct {
	repo             RecipeRepository
	ingredientMatcher *IngredientMatcher
	index           map[string][]int // keyword -> recipe IDs (for fast search)
	indexCh         chan int        // recipe ID to reindex (for background goroutine)
	mu              sync.RWMutex
}

// NewEnhancedSearchService creates an enhanced search service with ingredient matching
func NewEnhancedSearchService(repo RecipeRepository) *EnhancedSearchService {
	s := &EnhancedSearchService{
		repo:              repo,
		ingredientMatcher: NewIngredientMatcher(repo),
		index:            make(map[string][]int),
		indexCh:          make(chan int, 50),
	}
	go s.indexUpdater()
	s.rebuildIndex()
	return s
}

// indexUpdater runs in a goroutine and updates search index when recipes change
func (s *EnhancedSearchService) indexUpdater() {
	for id := range s.indexCh {
		s.reindexRecipe(id)
	}
}

func (s *EnhancedSearchService) reindexRecipe(recipeID int) {
	recipe, err := s.repo.GetByID(recipeID)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove old entries for this recipe
	for kw, ids := range s.index {
		newIds := make([]int, 0, len(ids))
		for _, id := range ids {
			if id != recipeID {
				newIds = append(newIds, id)
			}
		}
		if len(newIds) == 0 {
			delete(s.index, kw)
		} else {
			s.index[kw] = newIds
		}
	}

	// Add keywords from recipe name and description
	text := strings.ToLower(recipe.Name + " " + recipe.Description)
	words := strings.Fields(text)
	seen := make(map[string]bool)
	for _, w := range words {
		w = strings.Trim(w, ".,!?")
		if len(w) >= 2 && !seen[w] {
			seen[w] = true
			s.index[w] = append(s.index[w], recipeID)
		}
	}

	// Also index ingredient names for better search
	for _, ing := range recipe.Ingredients {
		ingName := strings.ToLower(ing.Ingredient.Name)
		if len(ingName) >= 2 && !seen[ingName] {
			seen[ingName] = true
			s.index[ingName] = append(s.index[ingName], recipeID)
		}
	}
}

func (s *EnhancedSearchService) rebuildIndex() {
	recipes := s.repo.GetAll()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index = make(map[string][]int)
	for _, recipe := range recipes {
		text := strings.ToLower(recipe.Name + " " + recipe.Description)
		words := strings.Fields(text)
		seen := make(map[string]bool)
		for _, w := range words {
			w = strings.Trim(w, ".,!?")
			if len(w) >= 2 && !seen[w] {
				seen[w] = true
				s.index[w] = append(s.index[w], recipe.ID)
			}
		}

		// Index ingredient names
		for _, ing := range recipe.Ingredients {
			ingName := strings.ToLower(ing.Ingredient.Name)
			if len(ingName) >= 2 && !seen[ingName] {
				seen[ingName] = true
				s.index[ingName] = append(s.index[ingName], recipe.ID)
			}
		}
	}
}

// NotifyRecipeChange notifies the indexer that a recipe was added or updated
func (s *EnhancedSearchService) NotifyRecipeChange(recipeID int) {
	select {
	case s.indexCh <- recipeID:
	default:
		// Channel full, skip
	}
}

// SearchByName returns recipes matching the query (uses repository search)
func (s *EnhancedSearchService) SearchByName(query string) []*models.Recipe {
	return s.repo.SearchByName(query)
}

// SearchByIngredients returns recipes that contain all given ingredients (exact match)
func (s *EnhancedSearchService) SearchByIngredients(names []string) []*models.Recipe {
	return s.repo.SearchByIngredients(names)
}

// AdvancedIngredientSearch performs sophisticated ingredient matching with scoring
func (s *EnhancedSearchService) AdvancedIngredientSearch(userIngredients []string, maxResults int) []RecipeMatchResult {
	return s.ingredientMatcher.MatchIngredients(userIngredients, maxResults)
}

// GetIngredientSubstitutes returns possible substitutes for a given ingredient
func (s *EnhancedSearchService) GetIngredientSubstitutes(ingredient string) []string {
	return s.ingredientMatcher.GetSubstitutes(ingredient)
}

// GetIngredientSynonyms returns synonyms for a given ingredient
func (s *EnhancedSearchService) GetIngredientSynonyms(ingredient string) []string {
	return s.ingredientMatcher.GetSynonyms(ingredient)
}

// AddIngredientSynonym allows adding custom synonyms at runtime
func (s *EnhancedSearchService) AddIngredientSynonym(canonical, synonym string) {
	s.ingredientMatcher.AddSynonym(canonical, synonym)
}

// AddIngredientSubstitute allows adding custom substitutes at runtime
func (s *EnhancedSearchService) AddIngredientSubstitute(ingredient, substitute string) {
	s.ingredientMatcher.AddSubstitute(ingredient, substitute)
}

// SearchRequest represents a comprehensive search request
type SearchRequest struct {
	Query         string   `json:"query,omitempty"`         // text search in name/description
	Ingredients   []string `json:"ingredients,omitempty"`   // ingredient matching
	MaxResults    int      `json:"max_results,omitempty"`   // limit results
	MinMatchScore float64  `json:"min_match_score,omitempty"` // minimum score threshold
	UseAdvanced   bool     `json:"use_advanced,omitempty"`   // use advanced matching
}

// SearchResponse represents a comprehensive search response
type SearchResponse struct {
	Recipes        []*models.Recipe      `json:"recipes,omitempty"`
	AdvancedMatches []RecipeMatchResult  `json:"advanced_matches,omitempty"`
	TotalCount     int                   `json:"total_count"`
	Query          string                `json:"query"`
	SearchType     string                `json:"search_type"`
}

// ComprehensiveSearch performs a comprehensive search based on the request
func (s *EnhancedSearchService) ComprehensiveSearch(req SearchRequest) SearchResponse {
	if req.MaxResults <= 0 {
		req.MaxResults = 50
	}

	var response SearchResponse
	response.Query = req.Query

	// Determine search type and perform appropriate search
	if req.UseAdvanced && len(req.Ingredients) > 0 {
		// Advanced ingredient matching
		matches := s.AdvancedIngredientSearch(req.Ingredients, req.MaxResults)

		// Filter by minimum score if specified
		if req.MinMatchScore > 0 {
			filtered := make([]RecipeMatchResult, 0)
			for _, match := range matches {
				if match.OverallScore >= req.MinMatchScore {
					filtered = append(filtered, match)
				}
			}
			matches = filtered
		}
		
		response.AdvancedMatches = matches
		response.TotalCount = len(matches)
		response.SearchType = "advanced_ingredient"
		
		// Also provide basic recipe list for compatibility
		response.Recipes = make([]*models.Recipe, len(matches))
		for i, match := range matches {
			response.Recipes[i] = match.Recipe
		}
		
	} else if len(req.Ingredients) > 0 {
		// Basic ingredient search (exact match)
		recipes := s.SearchByIngredients(req.Ingredients)
		if len(recipes) > req.MaxResults {
			recipes = recipes[:req.MaxResults]
		}
		response.Recipes = recipes
		response.TotalCount = len(recipes)
		response.SearchType = "basic_ingredient"
		
	} else if req.Query != "" {
		// Text search
		recipes := s.SearchByName(req.Query)
		if len(recipes) > req.MaxResults {
			recipes = recipes[:req.MaxResults]
		}
		response.Recipes = recipes
		response.TotalCount = len(recipes)
		response.SearchType = "text"
		
	} else {
		// Get all recipes
		recipes := s.repo.GetAll()
		if len(recipes) > req.MaxResults {
			recipes = recipes[:req.MaxResults]
		}
		response.Recipes = recipes
		response.TotalCount = len(recipes)
		response.SearchType = "all"
	}

	return response
}
