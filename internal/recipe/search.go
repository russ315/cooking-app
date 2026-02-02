package recipe

import (
	"strings"
	"sync"

	"cooking-app/internal/models"
	"cooking-app/internal/repository"
)

// SearchService encapsulates recipe search logic (Assignment 4 - Recipe Search Logic)
type SearchService struct {
	repo   *repository.RecipeRepository
	index  map[string][]int // keyword -> recipe IDs (for fast search)
	indexCh chan int        // recipe ID to reindex (for background goroutine)
	mu     sync.RWMutex
}

// NewSearchService creates a search service and starts background index updater (goroutine)
func NewSearchService(repo *repository.RecipeRepository) *SearchService {
	s := &SearchService{
		repo:    repo,
		index:   make(map[string][]int),
		indexCh: make(chan int, 50),
	}
	go s.indexUpdater()
	s.rebuildIndex()
	return s
}

// indexUpdater runs in a goroutine and updates search index when recipes change (Assignment 4 concurrency)
func (s *SearchService) indexUpdater() {
	for id := range s.indexCh {
		s.reindexRecipe(id)
	}
}

func (s *SearchService) reindexRecipe(recipeID int) {
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
}

func (s *SearchService) rebuildIndex() {
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
	}
}

// NotifyRecipeChange notifies the indexer that a recipe was added or updated (async via channel)
func (s *SearchService) NotifyRecipeChange(recipeID int) {
	select {
	case s.indexCh <- recipeID:
	default:
		// Channel full, skip
	}
}

// SearchByName returns recipes matching the query (uses repository search)
func (s *SearchService) SearchByName(query string) []*models.Recipe {
	return s.repo.SearchByName(query)
}

// SearchByIngredients returns recipes that contain all given ingredients
func (s *SearchService) SearchByIngredients(names []string) []*models.Recipe {
	return s.repo.SearchByIngredients(names)
}
