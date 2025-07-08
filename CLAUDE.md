# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **goscim** - a lightweight SCIM 2.0 (System for Cross-domain Identity Management) server implementation written in Go. It provides a REST API for managing user identities and groups with support for flexible schemas and extensions.

## Documentation Structure

- **docs/en/**: English documentation (installation, security, architecture, etc.)
- **docs/es/**: Spanish documentation (complete translations)
- **README.md**: Main project overview and community information
- **CONTRIBUTING.md**: Contribution guidelines
- **.github/**: Issue templates and PR templates for community engagement

## Development Commands

### Running the Application
```bash
# Start the SCIM server
go run .

# Set required environment variables
export SCIM_ADMIN_USER=Administrator
export SCIM_ADMIN_PASSWORD=admin!

# Optional environment variables
export SCIM_COUCHBASE_URL=localhost  # Default: localhost
export SCIM_PORT=:8080               # Default: :8080
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific test package
go test ./scim/parser -v

# Run tests with verbose output
go test -v ./...
```

### Dependencies
```bash
# Install/update dependencies
go mod tidy

# Download dependencies
go mod download
```

## Architecture Overview

### Core Components

1. **main.go** - Entry point that initializes the database, configures Gin router, and registers SCIM endpoints
2. **scim/config.go** - Configuration loader for schemas and resource types from JSON files
3. **scim/couchbase.go** - Couchbase database integration and bucket management
4. **scim/types.go** - Core SCIM data structures and types
5. **scim/parser/** - ANTLR-based SCIM filter query parser that converts SCIM filters to N1QL queries

### SCIM Operations (scim/op_*.go)
- **op_create.go** - POST operations for creating resources
- **op_read.go** - GET operations for retrieving individual resources
- **op_search.go** - GET operations for searching/listing resources with filtering
- **op_update.go** - PATCH operations for partial updates
- **op_replace.go** - PUT operations for full resource replacement
- **op_delete.go** - DELETE operations for removing resources

### Database Integration

- Uses **Couchbase** as the primary database
- Dynamic bucket creation based on resource types
- N1QL query generation for SCIM filter expressions
- Configurable bucket settings via JSON files in `config/bucketSettings/`

### Schema System

- **Flexible schema definition** via JSON files in `config/schemas/`
- **Resource type definitions** in `config/resourceType/`
- **Dynamic endpoint registration** based on resource types
- **Role-based access control** for read/write operations on attributes
- **Custom schema extensions** support (e.g., `urn:ietf:params:scim:schemas:extension:one:2.0:Element`)

### Key Features

1. **SCIM 2.0 Compliance** - Full CRUD operations, filtering, sorting, pagination
2. **Dynamic Schema Loading** - Add new resource types without code changes
3. **Filter Parser** - ANTLR-based parser converts SCIM filters to N1QL queries
4. **Role-based Security** - Attribute-level access control via `$reader` and `$writer` fields
5. **Flexible Configuration** - JSON-based configuration for schemas, resource types, and bucket settings

### REST API Structure

- **Base URL**: `/scim/v2`
- **Discovery endpoints**: `/ServiceProviderConfig`, `/ResourceTypes`, `/Schemas`
- **Resource endpoints**: Dynamically created based on resource type definitions
- **Standard SCIM operations**: Create, Read, Update, Delete, Search, Bulk (planned)

### Testing Strategy

- Unit tests in `scim/parser/parser_test.go` for filter parsing
- HTTP examples in `httpexamples/` directory for manual testing
- Uses REST Client format for API testing

## Development Notes

### Adding New Resource Types

1. Create schema JSON file in `config/schemas/`
2. Create resource type JSON file in `config/resourceType/`
3. Optionally create bucket configuration in `config/bucketSettings/`
4. Server will automatically register endpoints on startup

### Filter Parser Regeneration

When modifying `ScimFilter.g4`:
```bash
# Regenerate parser (requires ANTLR)
wget http://www.antlr.org/download/antlr-4.7-complete.jar
alias antlr='java -jar $PWD/antlr-4.7-complete.jar'
antlr -Dlanguage=Go -o scim/parser ScimFilter.g4
```

### Role-Based Access Control

- Use `$reader` and `$writer` arrays in schema attributes
- Roles are currently hardcoded but TODO: extract from authentication token
- Validation functions in `scim/validate_role.go`

### Couchbase Setup

For local development:
```bash
# Run Couchbase in Docker
docker run -d --name db -p 8091-8094:8091-8094 -p 11210:11210 couchbase

# Configure via web UI at http://localhost:8091/ui/index.html
# Create cluster with admin credentials matching environment variables
```