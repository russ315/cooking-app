# Ingredient Matching Algorithm

This document describes the enhanced ingredient matching algorithm implemented in the cooking app.

## Overview

The ingredient matching algorithm provides sophisticated recipe recommendations based on the ingredients users have available. It goes beyond simple exact matching to include:

- **Fuzzy matching** using Levenshtein distance
- **Synonym recognition** (e.g., "eggs" → "egg")
- **Ingredient substitution** suggestions
- **Scoring system** for recipe relevance

## Features

### 1. Ingredient Normalization

The algorithm normalizes ingredient names to handle variations:
- Plural/singular forms: "eggs" → "egg"
- Common aliases: "chicken breast" → "chicken"
- Descriptive terms: "all-purpose flour" → "flour"

### 2. Matching Types

#### Exact Match (Score: 1.0)
- Perfect ingredient name match after normalization

#### Synonym Match (Score: 0.9)
- Recognizes common ingredient synonyms
- Example: "tomatoes" matches "tomato"

#### Substitute Match (Score: 0.7)
- Suggests ingredient alternatives
- Example: "almond milk" can substitute "milk"

#### Fuzzy Match (Score: 0.6-1.0)
- Uses Levenshtein distance for similar names
- Handles typos and variations
- Example: "flour" partially matches "flower"

### 3. Recipe Scoring

The overall recipe score is calculated as:

```
base_score = (matched_ingredients / total_ingredients)
missing_penalty = (missing_ingredients * 0.2)
extra_penalty = (extra_ingredients * 0.05)
overall_score = base_score - missing_penalty - extra_penalty
```

## API Endpoints

### Advanced Ingredient Search
```http
POST /api/recipes/search/advanced
Content-Type: application/json

{
  "ingredients": ["egg", "flour", "milk"],
  "use_advanced": true,
  "max_results": 10,
  "min_match_score": 0.5
}
```

**Response:**
```json
{
  "search_type": "advanced_ingredient",
  "total_count": 2,
  "advanced_matches": [
    {
      "recipe": { ... },
      "overall_score": 0.95,
      "match_details": [
        {
          "ingredient": "egg",
          "score": 1.0,
          "match_type": "exact",
          "original": "egg"
        }
      ],
      "missing_count": 0,
      "extra_count": 1
    }
  ],
  "recipes": [...]
}
```

### Get Ingredient Substitutes
```http
GET /api/ingredients/egg/substitutes
```

**Response:**
```json
{
  "substitutes": ["flax egg", "chia egg", "apple sauce", "banana"]
}
```

### Get Ingredient Synonyms
```http
GET /api/ingredients/tomato/synonyms
```

**Response:**
```json
{
  "synonyms": ["tomatoes"]
}
```

### Add Custom Synonym (Protected)
```http
POST /api/ingredients/synonyms
Authorization: Bearer <token>
Content-Type: application/json

{
  "canonical": "chocolate",
  "synonym": "cocoa"
}
```

### Add Custom Substitute (Protected)
```http
POST /api/ingredients/substitutes
Authorization: Bearer <token>
Content-Type: application/json

{
  "ingredient": "butter",
  "substitute": "ghee"
}
```

## Usage Examples

### 1. Basic Ingredient Matching
```bash
curl -X POST http://localhost:8080/api/recipes/search/advanced \
  -H "Content-Type: application/json" \
  -d '{
    "ingredients": ["egg", "butter"],
    "use_advanced": true
  }'
```

### 2. Find Substitutes
```bash
curl http://localhost:8080/api/ingredients/milk/substitutes
```

### 3. Fuzzy Matching
The algorithm will find recipes even with misspelled ingredients:
```json
{
  "ingredients": ["eggs", "buttr"],  // "buttr" will fuzzy match with "butter"
  "use_advanced": true
}
```

## Built-in Data

The algorithm comes with pre-configured data for common cooking ingredients:

### Synonyms
- egg ↔ eggs
- tomato ↔ tomatoes
- garlic ↔ garlic cloves
- chicken ↔ chicken breast
- pasta ↔ spaghetti, penne, macaroni
- cheese ↔ cheddar, mozzarella, parmesan

### Substitutes
- egg → flax egg, chia egg, apple sauce, banana
- butter → margarine, coconut oil, vegetable oil
- milk → almond milk, soy milk, oat milk, coconut milk
- flour → almond flour, coconut flour, oat flour

## Algorithm Details

### Levenshtein Distance
The fuzzy matching uses the Levenshtein distance algorithm to calculate the edit distance between two strings. The similarity score is calculated as:

```
similarity = 1 - (edit_distance / max_length)
```

### Normalization Process
1. Convert to lowercase
2. Trim whitespace
3. Check alias mapping
4. Check synonym mapping
5. Return canonical form

### Performance Considerations
- Ingredient matching is O(n*m) where n is user ingredients and m is recipe ingredients
- Caching is used for synonym and substitute lookups
- Background indexing updates search terms when recipes change

## Testing

Run the comprehensive test suite:

```bash
go test ./internal/recipe/ -v
```

Tests cover:
- Ingredient normalization
- Similarity scoring
- Synonym and substitute matching
- Recipe matching with scoring
- Edge cases and error handling

## Future Enhancements

Potential improvements to consider:
1. **Machine Learning**: Train on user preferences to improve scoring
2. **Seasonal Ingredients**: Consider ingredient availability by season
3. **Dietary Restrictions**: Filter by vegan, gluten-free, etc.
4. **Cuisine Matching**: Consider cuisine type in matching
5. **Nutritional Analysis**: Match based on nutritional requirements
