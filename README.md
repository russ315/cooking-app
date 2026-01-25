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
│   ├── architecture.md  # System design description 
│   └── diagrams/        # Use-Case, ERD, and UML files 
├── .gitignore
├── go.mod               # Dependency management
└── README.md            # Project proposal and Gantt chart
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL (or your preferred database)

### Installation

```bash
go mod download
```

### Running the Application

```bash
go run cmd/cooking-app/main.go
```

The server will start on `http://localhost:8080`

## Development

### Project Proposal

[Add project proposal details here]

### Gantt Chart

[Add Gantt chart here]

## API Documentation

See `api/swagger.yaml` for detailed API documentation.

## Architecture

See `docs/architecture.md` for system design and architecture details.

## License

[Add license information]
