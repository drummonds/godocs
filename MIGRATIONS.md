# Database Migrations Guide

## Overview

godocs uses database migrations to manage schema changes. Migrations run automatically when the application starts and connects to the database.

## Migration Files

Migrations are located in `database/migrations/`:

- `000001_initial_schema.up.sql` - Creates the initial `documents` table
- `000002_add_fulltext_search.up.sql` - Adds full-text search capability
- `000003_add_word_cloud.up.sql` - Adds word frequency tracking

## Automatic Migration

When you start godocs, it automatically:
1. Connects to the database
2. Checks current migration version
3. Applies any pending migrations
4. Logs success or errors

```bash
# Migrations run automatically on startup
./godocs
# OR
./godocs-backend
```

## Common Issues

### Issue: "column full_text_search does not exist"

**Cause:** The full-text search migration didn't run on your existing database.

**Solution:** Run the manual migration script:

```bash
# Quick fix
./fix-search.sh

# Or manually with psql
psql -U godocs -d godocs -f fix-search-migration.sql
```

**Alternative:** Use environment variables if your database has different credentials:

```bash
POSTGRES_HOST=localhost \
POSTGRES_PORT=5432 \
POSTGRES_USER=godocs \
POSTGRES_DB=godocs \
./fix-search.sh
```

### Issue: "migration is dirty"

**Cause:** A previous migration failed midway.

**Fix:** The system will try to auto-recover, but if it doesn't:

```sql
-- Check migration status
SELECT * FROM schema_migrations;

-- Manually force to a known good version (e.g., version 3)
-- Only do this if you understand the implications!
```

### Issue: Migration errors on startup

**Symptoms:** Application fails to start with migration errors.

**Debugging:**
1. Check the logs for specific migration errors
2. Verify `database/migrations/` directory exists and is readable
3. Ensure database user has permission to create tables/indexes
4. Check that all .sql files are valid SQL

**Manual verification:**
```bash
# Test migration files syntax
psql -U godocs -d godocs -f database/migrations/000002_add_fulltext_search.up.sql
```

## Migration Version Tracking

The migration system uses a `schema_migrations` table:

```sql
-- Check current migration version
SELECT * FROM schema_migrations;

-- Should show version 3 (if all migrations applied)
```

## Creating New Migrations

If you need to add a new migration:

1. Create files with sequential numbers:
   ```
   database/migrations/000004_my_new_feature.up.sql
   database/migrations/000004_my_new_feature.down.sql
   ```

2. Write the `up` migration (what to apply):
   ```sql
   -- 000004_my_new_feature.up.sql
   ALTER TABLE documents ADD COLUMN new_field VARCHAR(255);
   ```

3. Write the `down` migration (how to rollback):
   ```sql
   -- 000004_my_new_feature.down.sql
   ALTER TABLE documents DROP COLUMN new_field;
   ```

4. Test the migration:
   ```bash
   # Start the application - migration runs automatically
   ./godocs
   ```

## Troubleshooting Commands

### Check if full-text search is set up

```sql
-- Check column exists
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'documents'
  AND column_name = 'full_text_search';

-- Check index exists
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'documents'
  AND indexname = 'idx_documents_full_text_search';

-- Check trigger exists
SELECT tgname, tgtype
FROM pg_trigger
WHERE tgname = 'trigger_update_full_text_search';

-- Test search works
SELECT COUNT(*) FROM documents
WHERE full_text_search @@ to_tsquery('english', 'test:*');
```

### Manually populate search index

If documents exist but search doesn't work:

```sql
-- Repopulate search index for all documents
UPDATE documents
SET full_text_search = to_tsvector('english', COALESCE(full_text, '') || ' ' || COALESCE(name, ''));

-- Check results
SELECT
    COUNT(*) as total,
    COUNT(full_text_search) as indexed
FROM documents;
```

### Reset migrations (DANGER!)

**⚠️ WARNING: This will drop and recreate your database!**

Only use in development:

```bash
# Backup first!
pg_dump -U godocs godocs > backup.sql

# Drop and recreate
psql -U postgres -c "DROP DATABASE godocs;"
psql -U postgres -c "CREATE DATABASE godocs OWNER godocs;"

# Migrations will run on next startup
./godocs
```

## Manual Migration via SQL

If automatic migrations aren't working, you can apply manually:

```bash
# Apply all migrations in order
for file in database/migrations/*up.sql; do
    echo "Applying $file..."
    psql -U godocs -d godocs -f "$file"
done
```

## Migration Best Practices

1. **Always test migrations** on a development database first
2. **Backup production data** before applying migrations
3. **Use IF NOT EXISTS** clauses to make migrations idempotent
4. **Keep migrations small** - one logical change per migration
5. **Never modify existing migrations** - create a new one instead
6. **Test both UP and DOWN** migrations

## Docker/Kubernetes Considerations

If running in containers:

```yaml
# Ensure migrations directory is mounted
volumes:
  - ./database/migrations:/app/database/migrations:ro

# Or copy into image
COPY database/migrations /app/database/migrations
```

## Getting Help

If migrations still aren't working:

1. Check application logs for specific errors
2. Verify database connection settings
3. Ensure database user has CREATE/ALTER permissions
4. Try the manual migration scripts in this directory
5. File an issue with:
   - Migration error messages
   - Database version (PostgreSQL/CockroachDB)
   - Output of `SELECT * FROM schema_migrations;`

## Quick Reference

| Task | Command |
|------|---------|
| Fix search migration | `./fix-search.sh` |
| Check migration status | `SELECT * FROM schema_migrations;` |
| Manually apply migration | `psql -U godocs -d godocs -f database/migrations/XXX.up.sql` |
| Check full-text search | `\d+ documents` (in psql) |
| Test search works | `curl "http://localhost:8000/api/search?term=test"` |
| Reindex all documents | `curl -X POST http://localhost:8000/api/search/reindex` |
