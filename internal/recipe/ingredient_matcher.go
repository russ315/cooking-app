package recipe

import (
	"math"
	"strings"
	"unicode"

	"cooking-app/internal/models"
)

// IngredientMatcher provides advanced ingredient matching capabilities
type IngredientMatcher struct {
	repo       RecipeRepository
	synonyms   map[string][]string // ingredient name -> list of synonyms
	aliases    map[string]string   // alias -> canonical name
	substitutes map[string][]string // ingredient -> possible substitutes
}

// NewIngredientMatcher creates a new ingredient matcher with predefined data
func NewIngredientMatcher(repo RecipeRepository) *IngredientMatcher {
	im := &IngredientMatcher{
		repo:       repo,
		synonyms:   make(map[string][]string),
		aliases:    make(map[string]string),
		substitutes: make(map[string][]string),
	}
	
	// Initialize ingredient synonyms and aliases
	im.initializeIngredientData()
	return im
}

// initializeIngredientData sets up common ingredient synonyms, aliases, and substitutes
func (im *IngredientMatcher) initializeIngredientData() {
	// Common cooking ingredient synonyms
	synonymData := map[string][]string{
		"egg":         {"eggs"},
		"flour":       {"all-purpose flour", "plain flour", "white flour", "wheat flour"},
		"sugar":       {"white sugar", "granulated sugar", "table sugar"},
		"butter":      {"unsalted butter", "salted butter"},
		"milk":        {"whole milk", "cow milk", "dairy milk"},
		"onion":       {"yellow onion", "white onion", "red onion"},
		"garlic":      {"garlic clove", "garlic cloves"},
		"tomato":      {"tomatoes", "fresh tomato", "ripe tomato"},
		"potato":      {"potatoes", "russet potato", "red potato"},
		"carrot":      {"carrots", "baby carrot"},
		"chicken":     {"chicken breast", "chicken thigh", "chicken meat"},
		"beef":        {"ground beef", "beef meat", "steak"},
		"rice":        {"white rice", "brown rice", "jasmine rice"},
		"pasta":       {"spaghetti", "penne", "macaroni"},
		"cheese":      {"cheddar", "mozzarella", "parmesan"},
		"olive oil":   {"extra virgin olive oil", "olive oil"},
		"salt":        {"table salt", "sea salt"},
		"pepper":      {"black pepper", "ground pepper"},
	}

	// Common aliases (alternative names)
	aliasData := map[string]string{
		"eggs":         "egg",
		"tomatoes":     "tomato",
		"potatoes":     "potato",
		"carrots":      "carrot",
		"onions":       "onion",
		"garlic cloves": "garlic",
		"chicken breast": "chicken",
		"ground beef":  "beef",
		"spaghetti":    "pasta",
		"cheddar":      "cheese",
		"mozzarella":   "cheese",
		"parmesan":     "cheese",
		"black pepper": "pepper",
		"sea salt":     "salt",
		"table salt":   "salt",
	}

	// Common substitutes
	substituteData := map[string][]string{
		"egg":         {"flax egg", "chia egg", "apple sauce", "banana"},
		"butter":      {"margarine", "coconut oil", "vegetable oil", "apple sauce"},
		"sugar":       {"honey", "maple syrup", "agave nectar", "stevia"},
		"milk":        {"almond milk", "soy milk", "oat milk", "coconut milk"},
		"flour":       {"almond flour", "coconut flour", "oat flour", "rice flour"},
		"sour cream":  {"yogurt", "buttermilk", "cream cheese"},
		"mayonnaise":  {"greek yogurt", "avocado", "hummus"},
		"rice":        {"quinoa", "couscous", "cauliflower rice"},
		"pasta":       {"zucchini noodles", "spaghetti squash", "rice noodles"},
	}

	// Load the data into maps
	for canonical, synonyms := range synonymData {
		im.synonyms[canonical] = synonyms
	}
	
	for alias, canonical := range aliasData {
		im.aliases[alias] = canonical
	}
	
	for ingredient, substitutes := range substituteData {
		im.substitutes[ingredient] = substitutes
	}
}

// normalizeIngredientName returns the canonical form of an ingredient name
func (im *IngredientMatcher) normalizeIngredientName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	
	// Check if it's an alias
	if canonical, exists := im.aliases[name]; exists {
		return canonical
	}
	
	// Check if it matches any synonym
	for canonical, synonyms := range im.synonyms {
		for _, synonym := range synonyms {
			if name == synonym {
				return canonical
			}
		}
	}
	
	return name
}

// levenshteinDistance calculates the edit distance between two strings
func (im *IngredientMatcher) levenshteinDistance(a, b string) int {
	a, b = strings.ToLower(a), strings.ToLower(b)
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = int(math.Min(float64(matrix[i-1][j]+1),
				math.Min(float64(matrix[i][j-1]+1), float64(matrix[i-1][j-1]+cost))))
		}
	}

	return matrix[len(a)][len(b)]
}

// similarityScore calculates a similarity score between two ingredient names (0-1)
func (im *IngredientMatcher) similarityScore(a, b string) float64 {
	a, b = strings.ToLower(a), strings.ToLower(b)
	
	// Exact match
	if a == b {
		return 1.0
	}
	
	// Check if one contains the other
	if strings.Contains(a, b) || strings.Contains(b, a) {
		shorter, longer := a, b
		if len(a) > len(b) {
			shorter, longer = b, a
		}
		return float64(len(shorter)) / float64(len(longer))
	}
	
	// Levenshtein distance similarity
	maxLen := math.Max(float64(len(a)), float64(len(b)))
	if maxLen == 0 {
		return 1.0
	}
	
	distance := float64(im.levenshteinDistance(a, b))
	similarity := 1.0 - (distance / maxLen)
	
	return math.Max(0, similarity)
}

// MatchResult represents a single ingredient match with its score
type MatchResult struct {
	Ingredient   string  `json:"ingredient"`
	Score        float64 `json:"score"`
	MatchType    string  `json:"match_type"` // "exact", "synonym", "fuzzy", "substitute"
	Original     string  `json:"original"`
}

// RecipeMatchResult represents a recipe with its overall match score and details
type RecipeMatchResult struct {
	Recipe       *models.Recipe  `json:"recipe"`
	OverallScore float64         `json:"overall_score"`
	MatchDetails []MatchResult   `json:"match_details"`
	MissingCount int             `json:"missing_count"`
	ExtraCount   int             `json:"extra_count"`
}

// MatchIngredients performs advanced ingredient matching against all recipes
func (im *IngredientMatcher) MatchIngredients(userIngredients []string, maxResults int) []RecipeMatchResult {
	// Normalize user ingredients
	normalizedUser := make(map[string]bool)
	for _, ing := range userIngredients {
		normalized := im.normalizeIngredientName(ing)
		if normalized != "" {
			normalizedUser[normalized] = true
		}
	}
	
	if len(normalizedUser) == 0 {
		return []RecipeMatchResult{}
	}
	
	// Get all recipes
	recipes := im.repo.GetAll()
	var results []RecipeMatchResult
	
	for _, recipe := range recipes {
		// Skip recipes with no ingredients
		if recipe.Ingredients == nil || len(recipe.Ingredients) == 0 {
			continue
		}
		
		matchResult := im.calculateRecipeMatch(recipe, normalizedUser, userIngredients)
		if matchResult.OverallScore > 0 {
			results = append(results, matchResult)
		}
	}
	
	// Sort by overall score (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].OverallScore > results[i].OverallScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	
	// Limit results
	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}
	
	return results
}

// calculateRecipeMatch calculates how well a recipe matches the user's ingredients
func (im *IngredientMatcher) calculateRecipeMatch(recipe *models.Recipe, userIngredients map[string]bool, originalUserIngredients []string) RecipeMatchResult {
	var matchDetails []MatchResult
	matchedRecipeIngredients := make(map[string]bool)
	matchedUserIngredients := make(map[string]bool)
	
	// Match each recipe ingredient against user ingredients
	for _, recipeIng := range recipe.Ingredients {
		recipeIngName := im.normalizeIngredientName(recipeIng.Ingredient.Name)
		bestMatch := im.findBestMatch(recipeIngName, originalUserIngredients)
		
		if bestMatch.Score > 0.3 { // Threshold for considering it a match
			matchDetails = append(matchDetails, bestMatch)
			matchedRecipeIngredients[recipeIngName] = true
			// Track the original ingredient that matched
			matchedUserIngredients[bestMatch.Original] = true
		}
	}
	
	// Calculate scores
	totalRecipeIngredients := len(recipe.Ingredients)
	matchedRecipeCount := len(matchedRecipeIngredients)
	missingCount := totalRecipeIngredients - matchedRecipeCount
	
	// Extra ingredients (user has but recipe doesn't need)
	// Count user ingredients that weren't matched to any recipe ingredient
	extraCount := 0
	for _, userIng := range originalUserIngredients {
		if !matchedUserIngredients[userIng] {
			extraCount++
		}
	}
	
	// Overall score calculation
	// Base score: percentage of ingredients matched
	baseScore := float64(matchedRecipeCount) / float64(totalRecipeIngredients)
	
	// Penalty for missing ingredients
	missingPenalty := float64(missingCount) * 0.2
	
	// Small penalty for extra ingredients (but less severe)
	extraPenalty := float64(extraCount) * 0.05
	
	overallScore := baseScore - missingPenalty - extraPenalty
	if overallScore < 0 {
		overallScore = 0
	}
	
	return RecipeMatchResult{
		Recipe:       recipe,
		OverallScore: overallScore,
		MatchDetails: matchDetails,
		MissingCount: missingCount,
		ExtraCount:   extraCount,
	}
}

// findBestMatch finds the best matching user ingredient for a recipe ingredient
func (im *IngredientMatcher) findBestMatch(recipeIngredient string, userIngredients []string) MatchResult {
	var bestMatch MatchResult
	
	for _, userIng := range userIngredients {
		normalizedUser := im.normalizeIngredientName(userIng)
		
		// Check exact match
		if normalizedUser == recipeIngredient {
			return MatchResult{
				Ingredient: recipeIngredient,
				Score:      1.0,
				MatchType:  "exact",
				Original:   userIng,
			}
		}
		
		// Check synonym match
		if im.isSynonym(normalizedUser, recipeIngredient) {
			return MatchResult{
				Ingredient: recipeIngredient,
				Score:      0.9,
				MatchType:  "synonym",
				Original:   userIng,
			}
		}
		
		// Check substitute match
		if im.isSubstitute(normalizedUser, recipeIngredient) {
			score := 0.7
			if score > bestMatch.Score {
				bestMatch = MatchResult{
					Ingredient: recipeIngredient,
					Score:      score,
					MatchType:  "substitute",
					Original:   userIng,
				}
			}
		}
		
		// Check fuzzy match
		similarity := im.similarityScore(normalizedUser, recipeIngredient)
		if similarity > 0.6 && similarity > bestMatch.Score {
			bestMatch = MatchResult{
				Ingredient: recipeIngredient,
				Score:      similarity,
				MatchType:  "fuzzy",
				Original:   userIng,
			}
		}
	}
	
	return bestMatch
}

// isSynonym checks if two ingredients are synonyms
func (im *IngredientMatcher) isSynonym(a, b string) bool {
	// Check direct synonym mapping
	if synonyms, exists := im.synonyms[a]; exists {
		for _, synonym := range synonyms {
			if synonym == b {
				return true
			}
		}
	}
	
	if synonyms, exists := im.synonyms[b]; exists {
		for _, synonym := range synonyms {
			if synonym == a {
				return true
			}
		}
	}
	
	return false
}

// isSubstitute checks if one ingredient can substitute another
func (im *IngredientMatcher) isSubstitute(a, b string) bool {
	if substitutes, exists := im.substitutes[a]; exists {
		for _, sub := range substitutes {
			if sub == b {
				return true
			}
		}
	}
	
	if substitutes, exists := im.substitutes[b]; exists {
		for _, sub := range substitutes {
			if sub == a {
				return true
			}
		}
	}
	
	return false
}

// GetSubstitutes returns possible substitutes for a given ingredient
func (im *IngredientMatcher) GetSubstitutes(ingredient string) []string {
	normalized := im.normalizeIngredientName(ingredient)
	if substitutes, exists := im.substitutes[normalized]; exists {
		return substitutes
	}
	return []string{}
}

// GetSynonyms returns synonyms for a given ingredient
func (im *IngredientMatcher) GetSynonyms(ingredient string) []string {
	normalized := im.normalizeIngredientName(ingredient)
	if synonyms, exists := im.synonyms[normalized]; exists {
		return synonyms
	}
	return []string{}
}

// AddSynonym allows adding custom synonyms at runtime
func (im *IngredientMatcher) AddSynonym(canonical, synonym string) {
	canonical = im.normalizeIngredientName(canonical)
	synonym = im.normalizeIngredientName(synonym)
	
	if canonical == "" || synonym == "" {
		return
	}
	
	if im.synonyms[canonical] == nil {
		im.synonyms[canonical] = []string{}
	}
	
	// Check if synonym already exists
	for _, existing := range im.synonyms[canonical] {
		if existing == synonym {
			return
		}
	}
	
	im.synonyms[canonical] = append(im.synonyms[canonical], synonym)
	im.aliases[synonym] = canonical
}

// AddSubstitute allows adding custom substitutes at runtime
func (im *IngredientMatcher) AddSubstitute(ingredient, substitute string) {
	ingredient = im.normalizeIngredientName(ingredient)
	substitute = im.normalizeIngredientName(substitute)
	
	if ingredient == "" || substitute == "" {
		return
	}
	
	if im.substitutes[ingredient] == nil {
		im.substitutes[ingredient] = []string{}
	}
	
	// Check if substitute already exists
	for _, existing := range im.substitutes[ingredient] {
		if existing == substitute {
			return
		}
	}
	
	im.substitutes[ingredient] = append(im.substitutes[ingredient], substitute)
}

// tokenize splits text into words, removing punctuation
func (im *IngredientMatcher) tokenize(text string) []string {
	var words []string
	var current strings.Builder
	
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
		}
	}
	
	if current.Len() > 0 {
		words = append(words, strings.ToLower(current.String()))
	}
	
	return words
}
