# digicert
# DigiCert

A RESTful API service for managing books built with Go, Chi router, and PostgreSQL.

## Features

- **Book Management**: Create, read, update, and delete books
- **Validation**: Request validation for book data (title, author, ISBN, published year)
- **Error Handling**: Comprehensive error handling with validation error responses
- **Testing**: Unit tests and integration tests with mocks
- **Database**: PostgreSQL with connection pooling
- **Clean Architecture**: Layered architecture (handler → service → repository)

## Project Structure

```
digicert/
├── internal/
│   ├── app/           # Application setup and configuration
│   ├── db/            # Database initialization and migrations
│   ├── handler/       # HTTP handlers
│   ├── model/         # Data models
│   ├── repo/          # Repository layer (data access)
│   └── service/       # Business logic layer
├── test/              # Integration tests
├── go.mod
├── go.sum
└── Readme.md
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/books` | List all books (with pagination) |
| GET | `/books/{id}` | Get a specific book |
| POST | `/books` | Create a new book |
| PUT | `/books/{id}` | Update a book |
| DELETE | `/books/{id}` | Delete a book |

## Request/Response Examples

### Create Book
**Request:**
```json
POST /books
{
  "title": "Go Programming",
  "author": "John Doe",
  "isbn": "978-1234567890",
  "published_year": 2020
}
```

**Response (201 Created):**
```json
{
  "id": "book-1",
  "title": "Go Programming",
  "author": "John Doe",
  "isbn": "978-1234567890",
  "published_year": 2020,
  "created_at": "2025-11-25T10:30:00Z",
  "updated_at": "2025-11-25T10:30:00Z",
  "version": 1
}
```

### Validation Errors
**Response (400 Bad Request):**
```json
{
  "title": "title is required",
  "author": "author must be at most 100 characters"
}
```

## Requirements

- Go 1.21+
- PostgreSQL 12+
- Chi v5 (HTTP router)
- pgx (PostgreSQL driver)
- testify (testing toolkit)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/praveen-anandh-jeyaraman/digicert.git
cd digicert
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=digicert
```

4. Run migrations (if available):
```bash
go run cmd/migrate/main.go
```

## Running the Application

```bash
go run cmd/main.go
```

The server will start on `http://localhost:8080`

## Running Tests

### Unit Tests (Handler layer)
```bash
go test ./internal/handler -v
```

### Integration Tests
```bash
go test ./test -v
```

### All Tests
```bash
go test ./... -v
```

## Validation Rules

- **Title**: Required, max 255 characters
- **Author**: Required, max 100 characters
- **Published Year**: Optional, must be between 1000 and current year + 1
- **ISBN**: Optional

## Error Handling

The API returns standardized error responses:

- **400 Bad Request**: Invalid input or validation errors
- **404 Not Found**: Resource not found
- **409 Conflict**: Version conflict or constraint violation
- **500 Internal Server Error**: Server-side errors

## Architecture

### Layers

1. **Handler Layer** (`internal/handler/`): HTTP request/response handling
2. **Service Layer** (`internal/service/`): Business logic and validation
3. **Repository Layer** (`internal/repo/`): Database operations
4. **Model Layer** (`internal/model/`): Data structures

### Dependency Injection

The application uses constructor-based dependency injection:

```go
svc := service.NewBookService(repo)
handler := handler.NewBookHandler(svc)
```
“The current domain fits well into a single relational table due to simple data relationships and low expected volume. If the system grew (e.g., multi-author support, publishers, inventory, user ratings), I would introduce additional normalized tables.

For extremely high throughput scenarios, scaling strategies like read replicas, partitioning, or sharding across tenants could be introduced — but these are not necessary for the current scope.”

## Author

Praveen Anandh Jeyaraman