# System Architecture

## Overview

This document describes the system design and architecture of the Cooking App.

## Components

### Authentication Module (`internal/auth/`)
- JWT token generation and validation
- User registration and login
- Password hashing and security

### Recipe Module (`internal/recipe/`)
- Recipe search and matching algorithms
- Ingredient-based recipe recommendations
- Recipe data management

### Inventory Module (`internal/inventory/`)
- User fridge/pantry management
- Ingredient tracking
- Inventory CRUD operations

### Database Module (`internal/db/`)
- Database connection management
- Migration scripts
- Database schema definitions




1. **Use-Case Diagram** (`use-case-diagram.pdf` or `.png`)
![usecase](/docs/diagrams/usecase.png)

2. **ERD (Entity Relationship Diagram)** (`erd.pdf` or `.png`)
![erd](/docs/diagrams/relationship.png)

3. **UML Diagrams** (`uml-diagrams.pdf` or `.png`)
![uml](/docs/diagrams/uml.png)

