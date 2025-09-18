# URL Shortener Service

A high-performance URL shortening service built with Go, Gin, and PostgreSQL. This service allows you to create short links, track their usage, and manage them through a REST API.

## Features

- ✅ Create short URLs with optional custom codes
- ✅ Set expiration times for URLs
- ✅ Track click statistics
- ✅ REST API with full CRUD operations
- ✅ PostgreSQL database with connection pooling
- ✅ Docker support with Docker Compose
- ✅ Rate limiting and CORS support
- ✅ Health check endpoint
- ✅ Comprehensive error handling

## Quick Start

### Using Docker Compose (Recommended)

1. Clone and navigate to the project:
```bash
git clone <your-repo>
cd url-shortener
```

2. Start the services:
```bash
docker-compose up -d
```

3. The service will be available at `http://localhost:8080`

### Manual Setup

1. Ensure you have Go 1.21+ and PostgreSQL installed

2. Create a PostgreSQL database:
```sql
CREATE DATABASE urlshortener;
```

3. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env with your database credentials
```

4. Install dependencies:
```bash
go mod download
```

5. Run database migrations:
```bash
# The init.sql file will be automatically executed when using Docker
# For manual setup, execute the SQL commands in init.sql
```

6. Run the application:
```bash
go run main.go
```

## API Documentation

### Base URL
```
http://localhost:8080
```

### Endpoints

#### Create Short URL
```http
POST /api/v1/urls
Content-Type: application/json

{
  "original_url": "https://example.com/very/long/url",
  "custom_code": "my-link",     // Optional: custom short code
  "expires_in": 24              // Optional: expiration in hours
}
```

**Response:**
```json
{
  "id": 1,
  "original_url": "https://example.com/very/long/url",
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "created_at": "2023-12-07T10:00:00Z",
  "expires_at": "2023-12-08T10:00:00Z"
}
```

#### Redirect to Original URL
```http
GET /{shortCode}
```
Returns a 301 redirect to the original URL and increments the click counter.

#### Get URL Statistics
```http
GET /api/v1/urls/{shortCode}/stats
```

**Response:**
```json
{
  "id": 1,
  "original_url": "https://example.com/very/long/url",
  "short_code": "abc123",
  "short_url": "http://localhost:8080/abc123",
  "created_at": "2023-12-07T10:00:00Z",
  "updated_at": "2023-12-07T10:30:00Z",
  "clicks": 15,
  "expires_at": "2023-12-08T10:00:00Z"
}
```

#### List URLs (Paginated)
```http
GET /api/v1/urls?page=1&per_page=10
```

**Response:**
```json
{
  "urls": [...],
  "total": 25,
  "page": 1,
  "per_page": 10,
  "has_next": true
}
```

#### Delete URL
```http
DELETE /api/v1/urls/{shortCode}
```

#### Health Check
```http
GET /health
```

## Configuration

The application uses environment variables for configuration. See `.env` for available options:

- `DB_HOST`: PostgreSQL host (default: localhost)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: PostgreSQL username (default: postgres)
- `DB_PASSWORD`: PostgreSQL password (default: postgres)
- `DB_NAME`: PostgreSQL database name (default: urlshortener)
- `DB_SSLMODE`: PostgreSQL SSL mode (default: disable)
- `SERVER_PORT`: Server port (default: 8080)
- `BASE_URL`: Base URL for generating short links (default: http://localhost:8080)

## Database Schema

The application uses a single `urls` table with the following structure:

```sql
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    original_url VARCHAR(2048) NOT NULL,
    short_code VARCHAR(20) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    clicks INTEGER DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE
);
```

## Development

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o url-shortener main.go
```

### Using Different Environments
The application automatically loads configuration from:
1. Environment variables
2. `.env` file (if present)

## Rate Limiting

The service includes basic rate limiting (60 requests per minute per IP). For production use, consider implementing a more sophisticated rate limiting solution using Redis.

## Security Considerations

- The service includes basic CORS support
- Custom short codes are validated for security
- URLs are validated before storage
- Rate limiting helps prevent abuse

## Monitoring

- Health check endpoint at `/health`
- Structured logging for all requests
- Database connection monitoring
- Click tracking for analytics

## License

This project is licensed under the MIT License - see the LICENSE file for details.