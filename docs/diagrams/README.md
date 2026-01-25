# Architecture Diagrams

This directory should contain the following diagrams from the assignment document:

## Required Diagrams

1. **Use-Case Diagram** (`use-case-diagram.pdf` or `.png`)
   - Shows actors and use cases for the Cooking App
   - Illustrates user interactions with the system

2. **ERD (Entity Relationship Diagram)** (`erd.pdf` or `.png`)
   - Database schema design
   - Shows relationships between entities (Users, Recipes, Ingredients, Inventory, etc.)

3. **UML Diagrams** (`uml-diagrams.pdf` or `.png`)
   - Class diagrams showing module structure
   - Sequence diagrams for key workflows
   - Component diagrams for system architecture

## Source Document

The original diagrams are located in:
`c:\Users\Ruslan\Documents\assign3.docx`

## Instructions

Please extract the diagrams from the Word document and save them in this directory as:
- PDF files (preferred for documentation)
- PNG files (for markdown inclusion)

You can:
1. Open the Word document
2. Export or save each diagram as an image/PDF
3. Place them in this `docs/diagrams/` directory
4. Update this README with the actual filenames

## Diagram Descriptions

### Use-Case Diagram
Should include actors such as:
- User (registered)
- Guest (unregistered)
- System

And use cases such as:
- Register/Login
- Manage Inventory
- Search Recipes
- Get Recommendations
- View Recipe Details

### ERD
Should show entities:
- Users
- Recipes
- Ingredients
- Inventory
- Recipe_Ingredients (junction table)
- etc.

### UML Diagrams
Should include:
- Class diagrams for each module (auth, recipe, inventory, db)
- Sequence diagrams for:
  - User registration flow
  - Recipe search flow
  - Inventory update flow
