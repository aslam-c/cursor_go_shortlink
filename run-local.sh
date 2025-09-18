#!/bin/bash

# Local setup script for URL Shortener
# This script helps you run the application without Docker

echo "🔗 URL Shortener - Local Setup"
echo "==============================="

# Check if PostgreSQL is running
echo "Checking PostgreSQL connection..."

# Set default environment variables if not set
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME=${DB_NAME:-urlshortener}
export DB_SSLMODE=${DB_SSLMODE:-disable}
export SERVER_PORT=${SERVER_PORT:-8080}
export BASE_URL=${BASE_URL:-http://localhost:8080}

echo "Configuration:"
echo "  Database: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "  Server: ${BASE_URL}"

# Build the application
echo ""
echo "Building application..."
go build -o url-shortener .

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "To run the application:"
    echo "1. Make sure PostgreSQL is running and the database '${DB_NAME}' exists"
    echo "2. Run the SQL commands from init.sql to create the schema"
    echo "3. Execute: ./url-shortener"
    echo ""
    echo "The service will be available at: ${BASE_URL}"
    echo ""
    echo "Alternative: Use Docker Compose for easier setup:"
    echo "  docker-compose up -d"
else
    echo "❌ Build failed!"
    exit 1
fi