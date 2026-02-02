# Cooking App

A smart recipe recommendation system that helps users find recipes based on ingredients they have available.

## Project Overview

This application allows users to:
- Manage their inventory (fridge/pantry items)
- Search for recipes based on available ingredients
- Get recipe recommendations and matching suggestions

## Project Structure

```
cooking-app/
├── cmd/
│   └── cooking-app/
│       └── main.go       # Entry point: initializes the server 
├── internal/
│   ├── auth/            # Auth logic (JWT, User registration) 
│   ├── recipe/          # Core logic for searching and matching 
│   ├── inventory/       # User fridge/pantry management 
│   └── db/              # Database connection and migrations 
├── api/
│   └── swagger.yaml     # API documentation 
├── docs/
│   ├── architecture.md      # System design description 
│   ├── project-proposal.md   # Project proposal document
│   ├── gantt-chart.md        # Project timeline and milestones
│   └── diagrams/             # Use-Case, ERD, and UML files 
├── .gitignore
├── go.mod               # Dependency management
└── README.md            # Project proposal and Gantt chart
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL

### PostgreSQL Setup

1. Install PostgreSQL and create a database:
   ```bash
   createdb cooking
   ```
2. Optional: set connection URL via environment variable:
   ```bash
   export DATABASE_URL="postgres://user:password@localhost:5432/cooking?sslmode=disable"
   ```
   Default (if unset): `postgres://postgres:postgres@localhost:5432/cooking?sslmode=disable`

3. On first run, the app creates tables and seeds sample data (ingredients, recipes, one user).

### Installation

```bash
go mod download
```

### Running the Application (Assignment 4)

From the project root:

```bash
go run .
```

The server will start on `http://localhost:8080`. You can also run `go run ./cmd/cooking-app` for the same behavior.

## Development

### Project Proposal

See `docs/project-proposal.md` for the complete project proposal.

### Gantt Chart

See `docs/gantt-chart.md` for the detailed project timeline and milestones.

## Assignment 4 – Milestone 2 (Core System)

This milestone implements:

- **Backend**: HTTP server (Gorilla Mux), JSON input/output, 10+ endpoints (user profile + recipe search).
- **Data model**: User, Recipe, Ingredient, RecipeIngredient (ERD). CRUD for User and Recipe. Thread-safe in-memory storage.
- **Recipe Search Logic**: Search by name (`GET /api/recipes?search=egg`), search by ingredients (`GET /api/recipes?ingredients=egg,flour`), list/get/create/update/delete recipes.
- **Concurrency**: Activity logger goroutine (channel-based); recipe search index updater goroutine (background reindex on recipe changes).

### Core features (Recipe Search)

| Feature | Endpoint | Description |
|--------|----------|-------------|
| List recipes | `GET /api/recipes` | All recipes |
| Search by name | `GET /api/recipes?search=...` | Recipes whose name/description contain the query |
| Search by ingredients | `GET /api/recipes?ingredients=egg,flour` | Recipes that contain all listed ingredients |
| Get recipe | `GET /api/recipes/{id}` | Recipe by ID |
| Create recipe | `POST /api/recipes` | Create recipe (JSON body) |
| Update recipe | `PUT /api/recipes/{id}` | Update recipe |
| Delete recipe | `DELETE /api/recipes/{id}` | Delete recipe |
| List ingredients | `GET /api/ingredients` | All ingredients |

## API Documentation

See `api/swagger.yaml` for detailed API documentation.

## Architecture

See `docs/architecture.md` for system design and architecture details.

