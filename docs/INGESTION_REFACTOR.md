# Ingestion Process Refactoring

## Overview

The document ingestion process has been refactored to use a step-based approach with hash verification and improved job tracking.

## Key Changes

### 1. Default Behavior Change
- **`INGRESS_DELETE` now defaults to `true`** - Source files are deleted after successful ingestion
- **No "done" folder by default** - Files are not moved to an archive folder
- **Backwards compatible** - Users can still set `INGRESS_DELETE=false` and `INGRESS_MOVE_FOLDER` if needed

### 2. Step-Based Ingestion Process

Each file now goes through three explicit steps with progress tracking:

#### Step 1: Hash and Database Record
- Calculate MD5 hash of source file
- Check for duplicates by hash
- Create initial database record with hash
- **Rollback**: Nothing to rollback (file unchanged)

#### Step 2: Move and Verify
- Copy file to documents folder
- Calculate hash of copied file
- Verify hash matches original
- Delete source file from ingress folder
- **Rollback**: Delete database record if verification fails

#### Step 3: Text Extraction and Indexing
- Extract text based on file type (PDF, images, text files)
- Update database record with full text
- PostgreSQL full-text search automatically indexed
- Register document view route
- **Fallback**: Store document even if text extraction fails

### 3. Hash Verification

**Before**: Hash calculated during database insert, file moved first
**After**: Hash calculated BEFORE moving, then verified AFTER moving

This ensures:
- No corrupted files in document storage
- Duplicate detection before processing
- Integrity verification at each step

### 4. Job Tracking Improvements

Jobs now show detailed per-file progress:

**Example Progress Messages:**
```
[1/5] invoice_2024.pdf - Step 1: Calculating hash
[1/5] invoice_2024.pdf - Step 2: Moving file
[1/5] invoice_2024.pdf - Step 3: Extracting text
```

**Result Format:**
```json
{
  "filesProcessed": 10,
  "filesTotal": 12,
  "duplicates": 2,
  "errors": 0
}
```

### 5. Benefits

✅ **More Robust**: Hash verification prevents corrupted files
✅ **No Duplicates**: Files without duplicates on disk (deleted from ingress)
✅ **Better Tracking**: See exactly which step a file is on
✅ **Easier Debugging**: Each step logged separately
✅ **Disk Space**: No "done" folder duplicating all ingested files
✅ **Rollback Support**: Failed steps can rollback changes

## Files Changed

### Backend
- `config/config.go` - Changed `INGRESS_DELETE` default to `true`
- `engine/ingestion_steps.go` (NEW) - Step-based ingestion functions
- `engine/engine.go` - Updated `ingressJobFuncWithTracking` to use new process
- `engine/engine.go` - Deprecated `ingressCleanup` function (kept for compatibility)

### Frontend
- `webapp/jobspage.go` - Added duplicate count display in job results

### Tests
- `engine/engine_test.go` - Updated to use temp directories and `INGRESS_DELETE=true`

## Migration Guide

### For Existing Users

**If you want the old behavior** (move to "done" folder):
```bash
export INGRESS_DELETE=false
export INGRESS_MOVE_FOLDER=/path/to/done
```

**If you want the new behavior** (delete source files):
```bash
# No changes needed - this is now the default
# Or explicitly set:
export INGRESS_DELETE=true
```

### Cleaning Up Old "done" Folder

If you have an existing `done/` folder with archived files:

```bash
# Check the size
du -sh done/

# If you don't need the archives, remove them
rm -rf done/
```

## API Changes

### Ingestion Endpoint Response

**Before:**
```json
{
  "message": "Ingestion completed",
  "scanned": 10,
  "errors": 0
}
```

**After:**
```json
{
  "message": "Ingestion started",
  "jobId": "01K8GSMYD7DF7G6KP82S3Y032G"
}
```

Use the Jobs API to track progress:
```bash
GET /api/jobs/{jobId}
GET /api/jobs/active
```

## Example Job Progress

```
2025-10-26 17:00:00 - Scanning ingress folder
2025-10-26 17:00:01 - [1/3] invoice.pdf - Step 1: Calculating hash
2025-10-26 17:00:02 - [1/3] invoice.pdf - Step 2: Moving file
2025-10-26 17:00:03 - [1/3] invoice.pdf - Step 3: Extracting text
2025-10-26 17:00:05 - [2/3] contract.pdf - Step 1: Calculating hash
2025-10-26 17:00:06 - Duplicate document detected, skipping
2025-10-26 17:00:07 - [3/3] report.pdf - Step 1: Calculating hash
2025-10-26 17:00:08 - [3/3] report.pdf - Step 2: Moving file
2025-10-26 17:00:09 - [3/3] report.pdf - Step 3: Extracting text
2025-10-26 17:00:10 - Updating word cloud
2025-10-26 17:00:11 - Complete: Processed 2 of 3 files (1 duplicate)
```

## Technical Details

### Hash Verification Flow

```
Source File (ingress/)
    ↓
Calculate Hash (MD5)
    ↓
Check Duplicate
    ↓
Create DB Record
    ↓
Copy to Documents
    ↓
Verify Hash Matches
    ↓
Delete Source File
    ↓
Extract Text
    ↓
Update DB + Search Index
```

### Error Handling

- **Step 1 Fails**: Nothing to clean up, source file remains in ingress
- **Step 2 Fails**: Delete DB record, source file remains in ingress
- **Step 3 Fails**: Document stored without text, can be reprocessed later

### Duplicate Detection

Files are considered duplicates if:
- MD5 hash matches existing document
- Source file is automatically deleted
- Counted in job results as "duplicates"
- No error - duplicate is expected behavior

## Testing

Run the test suite to verify the new process:

```bash
go test ./engine -v -run TestOCRProcessingAndDatabaseStorage
go test ./... -timeout 2m
```

All tests should pass with the new step-based ingestion.
