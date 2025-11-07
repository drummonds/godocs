package database

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

// runMigrations runs all Bun migrations
func (b *BunDB) runMigrations(ctx context.Context) error {
	// Create a simple migrations tracking table
	_, err := b.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bun_schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Check which migrations have been applied
	type AppliedMigration struct {
		bun.BaseModel `bun:"table:bun_schema_migrations"`
		Version       string `bun:"version"`
	}
	var applied []AppliedMigration
	err = b.db.NewSelect().
		Model(&applied).
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to check applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, m := range applied {
		appliedMap[m.Version] = true
	}

	// Run migrations in order
	migrations := []struct {
		version string
		name    string
		up      func(context.Context, *bun.DB) error
	}{
		{"001", "initial_schema", init001CreateDocumentsTable},
		{"002", "add_fulltext_search", init002AddFullTextSearch},
		{"003", "add_word_cloud", init003AddWordCloud},
		{"004", "create_jobs_table", init004CreateJobsTable},
	}

	for _, m := range migrations {
		if appliedMap[m.version] {
			continue
		}

		Logger.Info("Running migration", "version", m.version, "name", m.name)
		if err := m.up(ctx, b.db); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", m.version, err)
		}

		// Mark as applied
		_, err = b.db.NewInsert().
			Model(&AppliedMigration{Version: m.version}).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to mark migration %s as applied: %w", m.version, err)
		}
	}

	Logger.Info("All migrations completed successfully")
	return nil
}

// Migration 001: Create initial schema (documents and server_config tables)
func init001CreateDocumentsTable(ctx context.Context, db *bun.DB) error {
	Logger.Info("Running migration 001: Create initial schema")

	// Detect database dialect - check if it's PostgreSQL by checking dialect features
	_, isPostgres := db.Dialect().(interface{ SupportsReturning() bool })

	// Create documents table
	var createTableSQL string
	if isPostgres {
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS documents (
				id SERIAL PRIMARY KEY,
				name TEXT NOT NULL,
				path TEXT NOT NULL UNIQUE,
				ingress_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				folder TEXT NOT NULL,
				hash TEXT NOT NULL,
				ulid TEXT NOT NULL UNIQUE,
				document_type TEXT NOT NULL,
				full_text TEXT,
				url TEXT,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`
	} else {
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS documents (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				path TEXT NOT NULL UNIQUE,
				ingress_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				folder TEXT NOT NULL,
				hash TEXT NOT NULL,
				ulid TEXT NOT NULL UNIQUE,
				document_type TEXT NOT NULL,
				full_text TEXT,
				url TEXT,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`
	}

	_, err := db.ExecContext(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create documents table: %w", err)
	}

	// Create indexes for documents
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_documents_hash ON documents(hash)",
		"CREATE INDEX IF NOT EXISTS idx_documents_ulid ON documents(ulid)",
		"CREATE INDEX IF NOT EXISTS idx_documents_folder ON documents(folder)",
		"CREATE INDEX IF NOT EXISTS idx_documents_ingress_time ON documents(ingress_time DESC)",
	}

	for _, idx := range indexes {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create server_config table
	var createConfigSQL string
	var insertConfigSQL string
	if isPostgres {
		createConfigSQL = `
			CREATE TABLE IF NOT EXISTS server_config (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				listen_addr_ip TEXT DEFAULT '',
				listen_addr_port TEXT NOT NULL DEFAULT '8000',
				ingress_path TEXT NOT NULL DEFAULT '',
				ingress_delete BOOLEAN NOT NULL DEFAULT false,
				ingress_move_folder TEXT NOT NULL DEFAULT '',
				ingress_preserve BOOLEAN NOT NULL DEFAULT true,
				document_path TEXT NOT NULL DEFAULT '',
				new_document_folder TEXT DEFAULT '',
				new_document_folder_rel TEXT DEFAULT '',
				web_ui_pass BOOLEAN NOT NULL DEFAULT false,
				client_username TEXT DEFAULT '',
				client_password TEXT DEFAULT '',
				pushbullet_token TEXT DEFAULT '',
				tesseract_path TEXT DEFAULT '',
				use_reverse_proxy BOOLEAN NOT NULL DEFAULT false,
				base_url TEXT DEFAULT '',
				ingress_interval INTEGER NOT NULL DEFAULT 10,
				new_document_number INTEGER NOT NULL DEFAULT 5,
				server_api_url TEXT DEFAULT '',
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`
		insertConfigSQL = `INSERT INTO server_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING`
	} else {
		createConfigSQL = `
			CREATE TABLE IF NOT EXISTS server_config (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				listen_addr_ip TEXT DEFAULT '',
				listen_addr_port TEXT NOT NULL DEFAULT '8000',
				ingress_path TEXT NOT NULL DEFAULT '',
				ingress_delete BOOLEAN NOT NULL DEFAULT 0,
				ingress_move_folder TEXT NOT NULL DEFAULT '',
				ingress_preserve BOOLEAN NOT NULL DEFAULT 1,
				document_path TEXT NOT NULL DEFAULT '',
				new_document_folder TEXT DEFAULT '',
				new_document_folder_rel TEXT DEFAULT '',
				web_ui_pass BOOLEAN NOT NULL DEFAULT 0,
				client_username TEXT DEFAULT '',
				client_password TEXT DEFAULT '',
				pushbullet_token TEXT DEFAULT '',
				tesseract_path TEXT DEFAULT '',
				use_reverse_proxy BOOLEAN NOT NULL DEFAULT 0,
				base_url TEXT DEFAULT '',
				ingress_interval INTEGER NOT NULL DEFAULT 10,
				new_document_number INTEGER NOT NULL DEFAULT 5,
				server_api_url TEXT DEFAULT '',
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`
		insertConfigSQL = `INSERT OR IGNORE INTO server_config (id) VALUES (1)`
	}

	_, err = db.ExecContext(ctx, createConfigSQL)
	if err != nil {
		return fmt.Errorf("failed to create server_config table: %w", err)
	}

	// Insert default config row
	_, err = db.ExecContext(ctx, insertConfigSQL)
	if err != nil {
		return fmt.Errorf("failed to insert default config: %w", err)
	}

	Logger.Info("Migration 001 completed successfully")
	return nil
}

func init001RollbackDocumentsTable(ctx context.Context, db *bun.DB) error {
	Logger.Info("Rolling back migration 001")

	_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS server_config")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS documents")
	return err
}

// Migration 002: Add full-text search support
func init002AddFullTextSearch(ctx context.Context, db *bun.DB) error {
	Logger.Info("Running migration 002: Add full-text search")

	// Detect database dialect
	_, isPostgres := db.Dialect().(interface{ SupportsReturning() bool })

	if isPostgres {
		// PostgreSQL: Add tsvector column and GIN index
		_, err := db.ExecContext(ctx, `
			ALTER TABLE documents ADD COLUMN IF NOT EXISTS full_text_search tsvector
		`)
		if err != nil {
			Logger.Warn("Could not add full_text_search column (might already exist)", "error", err)
		}

		// Create GIN index for fast full-text searching
		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_documents_full_text_search ON documents USING GIN(full_text_search)
		`)
		if err != nil {
			return fmt.Errorf("failed to create full_text_search GIN index: %w", err)
		}

		// Create function to update search vector
		_, err = db.ExecContext(ctx, `
			CREATE OR REPLACE FUNCTION update_full_text_search()
			RETURNS TRIGGER AS $$
			BEGIN
				NEW.full_text_search = to_tsvector('english', COALESCE(NEW.full_text, '') || ' ' || COALESCE(NEW.name, ''));
				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql
		`)
		if err != nil {
			return fmt.Errorf("failed to create update_full_text_search function: %w", err)
		}

		// Create trigger to update search vector on insert/update
		_, err = db.ExecContext(ctx, `
			DROP TRIGGER IF EXISTS trigger_update_full_text_search ON documents
		`)
		if err != nil {
			Logger.Warn("Could not drop trigger (might not exist)", "error", err)
		}

		_, err = db.ExecContext(ctx, `
			CREATE TRIGGER trigger_update_full_text_search
				BEFORE INSERT OR UPDATE OF full_text, name ON documents
				FOR EACH ROW
				EXECUTE FUNCTION update_full_text_search()
		`)
		if err != nil {
			return fmt.Errorf("failed to create trigger: %w", err)
		}

		// Update existing documents to populate the search vector
		_, err = db.ExecContext(ctx, `
			UPDATE documents
			SET full_text_search = to_tsvector('english', COALESCE(full_text, '') || ' ' || COALESCE(name, ''))
		`)
		if err != nil {
			Logger.Warn("Could not update existing documents (table might be empty)", "error", err)
		}
	} else {
		// SQLite: Add a simple full_text_search column for LIKE queries
		_, err := db.ExecContext(ctx, `
			ALTER TABLE documents ADD COLUMN full_text_search TEXT
		`)
		if err != nil {
			// Column might already exist, ignore error
			Logger.Warn("Could not add full_text_search column (might already exist)", "error", err)
		}

		// Create index for faster LIKE queries
		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_documents_full_text_search ON documents(full_text_search)
		`)
		if err != nil {
			return fmt.Errorf("failed to create full_text_search index: %w", err)
		}
	}

	Logger.Info("Migration 002 completed successfully")
	return nil
}

func init002RollbackFullTextSearch(ctx context.Context, db *bun.DB) error {
	Logger.Info("Rolling back migration 002")

	// SQLite doesn't support DROP COLUMN easily, so we skip it
	// The column will remain but won't be used

	Logger.Info("Migration 002 rollback completed (column retained for SQLite compatibility)")
	return nil
}

// Migration 003: Add word cloud tables
func init003AddWordCloud(ctx context.Context, db *bun.DB) error {
	Logger.Info("Running migration 003: Add word cloud tables")

	// Detect database dialect
	_, isPostgres := db.Dialect().(interface{ SupportsReturning() bool })

	// Create word_frequencies table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS word_frequencies (
			word TEXT PRIMARY KEY,
			frequency INTEGER DEFAULT 1,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create word_frequencies table: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_word_frequencies_frequency ON word_frequencies(frequency DESC)",
		"CREATE INDEX IF NOT EXISTS idx_word_frequencies_updated ON word_frequencies(last_updated DESC)",
	}

	for _, idx := range indexes {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create word_cloud_metadata table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS word_cloud_metadata (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			last_full_calculation TIMESTAMP,
			total_documents_processed INTEGER DEFAULT 0,
			total_words_indexed INTEGER DEFAULT 0,
			version INTEGER DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create word_cloud_metadata table: %w", err)
	}

	// Insert default metadata row
	var insertMetadataSQL string
	if isPostgres {
		insertMetadataSQL = `INSERT INTO word_cloud_metadata (id) VALUES (1) ON CONFLICT (id) DO NOTHING`
	} else {
		insertMetadataSQL = `INSERT OR IGNORE INTO word_cloud_metadata (id) VALUES (1)`
	}

	_, err = db.ExecContext(ctx, insertMetadataSQL)
	if err != nil {
		return fmt.Errorf("failed to insert default metadata: %w", err)
	}

	Logger.Info("Migration 003 completed successfully")
	return nil
}

func init003RollbackWordCloud(ctx context.Context, db *bun.DB) error {
	Logger.Info("Rolling back migration 003")

	_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS word_cloud_metadata")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS word_frequencies")
	return err
}

// Migration 004: Create jobs table
func init004CreateJobsTable(ctx context.Context, db *bun.DB) error {
	Logger.Info("Running migration 004: Create jobs table")

	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			progress INTEGER DEFAULT 0,
			current_step TEXT DEFAULT '',
			total_steps INTEGER DEFAULT 0,
			message TEXT DEFAULT '',
			error TEXT,
			result TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create jobs table: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status)",
		"CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs(type)",
		"CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_jobs_completed_at ON jobs(completed_at) WHERE completed_at IS NOT NULL",
	}

	for _, idx := range indexes {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			// Partial indexes might not be supported in all SQLite versions
			Logger.Warn("Could not create index (might not be supported)", "error", err)
		}
	}

	Logger.Info("Migration 004 completed successfully")
	return nil
}

func init004RollbackJobsTable(ctx context.Context, db *bun.DB) error {
	Logger.Info("Rolling back migration 004")

	_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS jobs")
	return err
}
