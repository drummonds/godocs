-- Create jobs table for tracking background operations
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    progress INTEGER NOT NULL DEFAULT 0,
    current_step TEXT NOT NULL DEFAULT '',
    total_steps INTEGER NOT NULL DEFAULT 0,
    message TEXT NOT NULL DEFAULT '',
    error TEXT,
    result TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Create index for faster queries by status
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);

-- Create index for faster queries by type
CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs(type);

-- Create index for faster queries by created_at (for pagination)
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);

-- Create index for cleaning old jobs
CREATE INDEX IF NOT EXISTS idx_jobs_completed_at ON jobs(completed_at) WHERE completed_at IS NOT NULL;

-- Add comment to table
COMMENT ON TABLE jobs IS 'Tracks background jobs and operations with progress';

-- Add comments to columns
COMMENT ON COLUMN jobs.id IS 'ULID of the job';
COMMENT ON COLUMN jobs.type IS 'Type of job: ingestion, cleanup, wordcloud, search_reindex';
COMMENT ON COLUMN jobs.status IS 'Current status: pending, running, completed, failed, cancelled';
COMMENT ON COLUMN jobs.progress IS 'Progress percentage (0-100)';
COMMENT ON COLUMN jobs.current_step IS 'Human-readable description of current step';
COMMENT ON COLUMN jobs.total_steps IS 'Total number of steps in the job';
COMMENT ON COLUMN jobs.message IS 'Status message or description';
COMMENT ON COLUMN jobs.error IS 'Error message if job failed';
COMMENT ON COLUMN jobs.result IS 'JSON result data when job completes';
