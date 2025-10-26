-- Manual fix for missing full_text_search column
-- Run this if you get "column full_text_search does not exist" error
-- psql -U goedms -d goedms -f fix-search-migration.sql

-- Check if column exists first
DO $$
BEGIN
    -- Add a tsvector column for full-text search
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='documents' AND column_name='full_text_search'
    ) THEN
        ALTER TABLE documents ADD COLUMN full_text_search tsvector;
        RAISE NOTICE 'Added full_text_search column';
    ELSE
        RAISE NOTICE 'full_text_search column already exists';
    END IF;
END $$;

-- Create a GIN index for fast full-text searching
CREATE INDEX IF NOT EXISTS idx_documents_full_text_search ON documents USING GIN(full_text_search);

-- Create a function to automatically update the search vector when full_text changes
CREATE OR REPLACE FUNCTION update_full_text_search()
RETURNS TRIGGER AS $$
BEGIN
    NEW.full_text_search = to_tsvector('english', COALESCE(NEW.full_text, '') || ' ' || COALESCE(NEW.name, ''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to update search vector on insert/update
DROP TRIGGER IF EXISTS trigger_update_full_text_search ON documents;
CREATE TRIGGER trigger_update_full_text_search
    BEFORE INSERT OR UPDATE OF full_text, name ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_full_text_search();

-- Update existing documents to populate the search vector
UPDATE documents SET full_text_search = to_tsvector('english', COALESCE(full_text, '') || ' ' || COALESCE(name, ''));

-- Show results
SELECT
    COUNT(*) as total_documents,
    COUNT(full_text_search) as documents_with_search_index,
    COUNT(*) - COUNT(full_text_search) as documents_without_index
FROM documents;

\echo 'Full-text search migration completed successfully!'
