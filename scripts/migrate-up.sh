#!/bin/bash

# Run database migrations up

set -e

# Default database URL
DB_URL="${DATABASE_URL:-postgres://hruser:hrpassword@localhost:5432/hrmanagement?sslmode=disable}"

echo "Running database migrations..."
echo "Database URL: $DB_URL"

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo "migrate is not installed. Installing..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Run migrations
migrate -path internal/database/migrations -database "$DB_URL" up

echo "✅ Database migrations completed successfully!"