# Assignment 4 – Milestone 2: Implementation Notes

## How This Implementation Follows Assignment 3 Design

1. **Architecture (Assignment 3)**  
   The design uses `internal/recipe/` for search logic, `internal/repository` for storage, and `internal/handler` for HTTP. This codebase keeps that layout:
   - **Recipe search logic** lives in `internal/recipe/search.go` (search by name, by ingredients, background index).
   - **Domain models** in `internal/models/` match the ERD: `Recipe`, `Ingredient`, `RecipeIngredient` (many-to-many).
   - **Persistence** is in-memory in `internal/repository/recipe_repository.go` with mutex for safe concurrent access.
   - **HTTP** is in `internal/handler/recipe_handler.go` with JSON in/out.

2. **ERD alignment**  
   - **Recipe**: id, name, description, instructions, prep_time_min, cook_time_min, created_at.  
   - **Ingredient**: id, name.  
   - **RecipeIngredient**: recipe_id, ingredient_id, quantity (junction table).  
   All are represented in `internal/models/recipe.go` and used in the repository and handlers.

3. **Core features (3–5)**  
   Implemented recipe-related features:
   - List recipes (`GET /api/recipes`).
   - Search by name (`GET /api/recipes?search=...`).
   - Search by ingredients (`GET /api/recipes?ingredients=egg,flour`).
   - Get recipe by ID (`GET /api/recipes/{id}`).
   - Create recipe (`POST /api/recipes`).  
   Plus user profile CRUD from the existing code.

4. **Technical requirements**
   - **Backend**: HTTP server (Gorilla Mux), 3+ endpoints, JSON in/out.
   - **Data model & storage**: Structures match ERD; CRUD for Recipe (and User); thread-safe access (mutex in repositories).
   - **Concurrency**: Activity logger goroutine (channel-based); recipe search index updater goroutine (background reindex on create/update).

## Running and Demo

- From project root: `go run .`
- Server: `http://localhost:8080`
- Try in Postman or browser:
  - `GET /api/recipes` – list all recipes
  - `GET /api/recipes?search=egg` – search by name
  - `GET /api/recipes?ingredients=egg,flour` – search by ingredients
  - `GET /api/recipes/1` – get recipe 1
  - `GET /api/ingredients` – list ingredients

## Recipe Search Logic (Topic)

- **Search by name**: Substring match (case-insensitive) on recipe name and description in `RecipeRepository.SearchByName`.
- **Search by ingredients**: Recipes that contain *all* given ingredients (by name) in `RecipeRepository.SearchByIngredients`.
- **Background index**: `SearchService` maintains a keyword index and updates it in a goroutine when recipes are created/updated (`NotifyRecipeChange` → channel → `indexUpdater`).
