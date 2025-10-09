#!/bin/bash

# Rollback database migrations

set -e

# Default database URL
DB_URL="${DATABASE_URL:-postgres://hruser:hrpassword@localhost:5432/hrmanagement?sslmode=disable}"

echo "Rolling back database migrations..."
echo "Database URL: $DB_URL"

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo "migrate is not installed. Please install it first."
    echo "Run: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Get current migration version
VERSION=$(migrate -path internal/database/migrations -database "$DB_URL" version 2>/dev/null || echo "No migrations applied")
echo "Current migration version: $VERSION"

# Confirm rollback
read -p "Are you sure you want to rollback migrations? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Migration rollback cancelled."
    exit 0
fi

# Rollback migrations
migrate -path internal/database/migrations -database "$DB_URL" down

echo "✅ Database migrations rolled back successfully!"
