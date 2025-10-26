#!/bin/bash
# Fix missing full_text_search column

echo "=================================="
echo "goEDMS Search Fix Script"
echo "=================================="
echo ""
echo "This script will add the full_text_search column to your database"
echo "if it's missing, and populate it for all existing documents."
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "ERROR: psql command not found. Please install PostgreSQL client tools."
    exit 1
fi

# Default connection parameters (can be overridden with environment variables)
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-goedms}"
DB_NAME="${POSTGRES_DB:-goedms}"

echo "Connection details:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo ""
echo "Press Enter to continue, or Ctrl+C to cancel..."
read

# Run the migration
echo "Running migration..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f fix-search-migration.sql

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Migration completed successfully!"
    echo ""
    echo "You can now use search. Try:"
    echo "  curl http://localhost:8000/api/search?term=test"
    echo ""
else
    echo ""
    echo "❌ Migration failed. Check the error messages above."
    echo ""
    exit 1
fi
