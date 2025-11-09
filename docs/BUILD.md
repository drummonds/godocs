# Building godocs

This document describes how to build godocs from source.

## Prerequisites

- Go 1.22 or later
- Git (for version tagging)
- Node.js/npm (optional, for development tools)

## Quick Build

### Build Frontend (WebAssembly)

```bash
./build-wasm.sh
```

This will:
- Automatically detect the version from git tags
- Embed the version and build date into the binary
- Create `web/app.wasm` with version information

### Build Backend

```bash
go build -o godocs main.go
```

Or for a specific output directory:

```bash
mkdir -p build
go build -o build/godocs main.go
```

### Build Both (Using Taskfile)

If you have [Task](https://taskfile.dev/) installed:

```bash
task build
```

This will build both the WASM frontend and backend binary.

## Manual Build with Version Information

### Frontend (WASM)

```bash
VERSION=$(git describe --tags --always)
BUILD_DATE=$(date -u +"%Y-%m-%d")

GOOS=js GOARCH=wasm go build \
  -ldflags="-X 'github.com/drummonds/godocs/webapp.Version=$VERSION' \
            -X 'github.com/drummonds/godocs/webapp.BuildDate=$BUILD_DATE'" \
  -o web/app.wasm \
  ./cmd/webapp
```

### Backend

```bash
VERSION=$(git describe --tags --always)
BUILD_DATE=$(date -u +"%Y-%m-%d")

go build \
  -ldflags="-X 'main.Version=$VERSION' \
            -X 'main.BuildDate=$BUILD_DATE'" \
  -o godocs \
  main.go
```

## Version Information

The version is automatically determined from git:

```bash
git describe --tags --always
# Output example: v0.7.0-5-ge774e6d2
#   v0.7.0 = last tag
#   5 = commits since tag
#   ge774e6d2 = current commit hash
```

This version appears in:
- Frontend navbar: `v0.7.0-5-ge774e6d2 | 2025-10-26`
- About page
- API responses

## Development Builds

For development, you can build without version information (will show "dev"):

```bash
# Frontend
GOOS=js GOARCH=wasm go build -o web/app.wasm ./cmd/webapp

# Backend
go build -o godocs main.go
```

## Using Taskfile

The project includes a Taskfile for common build tasks:

```bash
# List all available tasks
task --list

# Build everything
task build

# Build only WASM frontend
task build:wasm

# Build only backend
task build:backend

# Run tests
task test

# Clean build artifacts
task clean

# Generate OpenAPI docs
task openapi
```

## Release Builds

For official releases:

1. **Tag the release:**
   ```bash
   git tag -a v0.8.0 -m "Release v0.8.0"
   git push origin v0.8.0
   ```

2. **Build with the new tag:**
   ```bash
   ./build-wasm.sh
   task build
   ```

3. **Verify version:**
   ```bash
   strings web/app.wasm | grep "v0.8.0"
   ./build/godocs --version  # (if version flag is implemented)
   ```

## Troubleshooting

### "dev" Appears as Version

If you see "dev" in the frontend:
- Make sure you built with `./build-wasm.sh` or `task build:wasm`
- Check if git is available: `git --version`
- Verify git tags exist: `git describe --tags --always`

### WASM File Not Found

Make sure the WASM file is in the correct location:
```bash
ls -lh web/app.wasm
```

### Version Doesn't Update

Clear browser cache or do a hard refresh (Ctrl+F5) to reload the WASM file.

## Build Artifacts

After building, you'll have:

```
build/
  └── godocs          # Backend binary

web/
  ├── app.wasm        # Frontend WASM (7.2MB)
  └── wasm_exec.js    # WASM runtime

docs/
  ├── openapi.yaml    # API documentation
  └── swagger.yaml    # Swagger docs
```

## CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Build
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Needed for git describe

      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Build
        run: |
          ./build-wasm.sh
          go build -o godocs main.go

      - name: Test
        run: go test ./...
```

## See Also

- [INGESTION_REFACTOR.md](INGESTION_REFACTOR.md) - Recent changes to ingestion process
- [DIAGRAMS.md](DIAGRAMS.md) - System architecture diagrams
- [README.md](../README.md) - Main project documentation
