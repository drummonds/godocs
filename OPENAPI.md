# OpenAPI Specification Generation for godocs

## Overview

godocs currently doesn't have an OpenAPI specification, but there are several excellent tools to generate one from the Go source code.

## Recommended: Swaggo (swag)

**Best for:** Echo/Gin/Fiber applications, most popular Go OpenAPI tool

### Installation

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### How It Works

Add annotations to your route handlers:

```go
// GetDocuments retrieves the latest documents
// @Summary Get latest documents
// @Description Get a paginated list of the most recently ingested documents
// @Tags documents
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Success 200 {array} Document
// @Failure 500 {object} map[string]string
// @Router /api/documents/latest [get]
func (serverHandler *ServerHandler) GetLatestDocuments(c echo.Context) error {
    // ... implementation
}
```

### Generate Spec

```bash
# Initialize (first time only)
swag init

# Generate docs
swag init -g main.go -o ./docs --parseDependency --parseInternal

# Output: docs/swagger.json and docs/swagger.yaml
```

### Pros
- ✅ Well-maintained, large community
- ✅ Works great with Echo
- ✅ Annotations keep docs close to code
- ✅ Generates Swagger UI automatically
- ✅ Supports Echo, Gin, Fiber, net/http

### Cons
- ❌ Requires adding annotations to every endpoint
- ❌ Can clutter code with comments

---

## Alternative: go-swagger

**Best for:** Contract-first development, sophisticated validation

### Installation

```bash
go install github.com/go-swagger/go-swagger/cmd/swagger@latest
```

### How It Works

Uses different annotation style:

```go
// swagger:route GET /api/documents/latest documents listDocuments
//
// List latest documents
//
// Returns a list of recently ingested documents
//
// Responses:
//   200: documentsResponse
//   500: errorResponse
func (serverHandler *ServerHandler) GetLatestDocuments(c echo.Context) error {
    // ... implementation
}
```

### Generate Spec

```bash
swagger generate spec -o swagger.yaml --scan-models
```

### Pros
- ✅ More powerful code generation
- ✅ Strong validation support
- ✅ Can generate server/client code

### Cons
- ❌ Different annotation syntax
- ❌ Less Echo-specific documentation
- ❌ Steeper learning curve

---

## Alternative: Manual OpenAPI

**Best for:** Small APIs, custom requirements, existing specs

### Create openapi.yaml manually

```yaml
openapi: 3.0.0
info:
  title: godocs API
  version: 1.0.0
  description: Document Management System API

servers:
  - url: http://localhost:8000
    description: Local development
  - url: http://backend:8000
    description: Backend server

paths:
  /api/documents/latest:
    get:
      summary: Get latest documents
      tags:
        - documents
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Document'

components:
  schemas:
    Document:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
        path:
          type: string
```

### Pros
- ✅ Complete control over spec
- ✅ No code annotations needed
- ✅ Can import from other tools
- ✅ Language agnostic

### Cons
- ❌ Manual maintenance
- ❌ Can get out of sync with code
- ❌ Time-consuming for large APIs

---

## Alternative: oapi-codegen

**Best for:** Spec-first development (opposite direction)

### Installation

```bash
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
```

### How It Works

1. Write OpenAPI spec first
2. Generate Go code from spec

```bash
oapi-codegen -package api -generate types,server,spec openapi.yaml > api/generated.go
```

### Pros
- ✅ Ensures API matches spec
- ✅ Type-safe code generation
- ✅ Good for contract-first approach

### Cons
- ❌ Wrong direction for existing code
- ❌ Requires rewriting existing handlers

---

## Comparison Table

| Tool | Maintenance | Learning Curve | Echo Support | Community | Output |
|------|-------------|----------------|--------------|-----------|--------|
| **swag** | Low | Easy | ⭐⭐⭐⭐⭐ | Large | OpenAPI 2.0/3.0 |
| **go-swagger** | Medium | Medium | ⭐⭐⭐ | Large | OpenAPI 2.0 |
| **Manual** | High | Easy | N/A | N/A | Any version |
| **oapi-codegen** | Low | Medium | ⭐⭐⭐⭐ | Medium | Generates Go |

---

## Recommended Approach for godocs

### Phase 1: Quick Start with Swag

1. Install swag:
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

2. Add general API info to `main.go`:
   ```go
   // @title godocs API
   // @version 1.0
   // @description Electronic Document Management System API
   // @host localhost:8000
   // @BasePath /
   func main() {
       // ... existing code
   }
   ```

3. Annotate a few key endpoints in `engine/routes.go`

4. Generate spec:
   ```bash
   swag init -g main.go -o ./docs
   ```

5. Serve the spec:
   ```go
   // Add to main.go
   import "github.com/swaggo/echo-swagger"

   e.GET("/swagger/*", echoSwagger.WrapHandler)
   ```

### Phase 2: Complete Documentation

Gradually add annotations to all endpoints as you work on them.

### Phase 3: CI/CD Integration

Add to your build pipeline:
```yaml
# .github/workflows/api-docs.yml
- name: Generate OpenAPI Spec
  run: |
    go install github.com/swaggo/swag/cmd/swag@latest
    swag init -g main.go -o ./docs

- name: Upload API Spec
  uses: actions/upload-artifact@v2
  with:
    name: openapi-spec
    path: docs/swagger.yaml
```

---

## Example: Annotating godocs Endpoints

### Before (current code):
```go
func (serverHandler *ServerHandler) SearchDocuments(context echo.Context) error {
	searchTerm := context.QueryParams().Get("term")
	documents, err := serverHandler.DB.SearchDocuments(searchTerm)
	// ...
	return context.JSON(http.StatusOK, fullResults)
}
```

### After (with swag annotations):
```go
// SearchDocuments searches for documents by term
// @Summary Search documents
// @Description Performs full-text search across all documents using PostgreSQL
// @Tags search
// @Accept json
// @Produce json
// @Param term query string true "Search term"
// @Success 200 {array} fileTreeStruct "Search results"
// @Success 204 "No results found"
// @Failure 404 {object} string "Empty search term"
// @Failure 500 {object} error "Internal server error"
// @Router /api/search [get]
func (serverHandler *ServerHandler) SearchDocuments(context echo.Context) error {
	searchTerm := context.QueryParams().Get("term")
	documents, err := serverHandler.DB.SearchDocuments(searchTerm)
	// ... implementation unchanged
	return context.JSON(http.StatusOK, fullResults)
}
```

---

## Task Integration

Add to your `Taskfile.yml`:

```yaml
  # OpenAPI tasks
  openapi:generate:
    desc: Generate OpenAPI specification
    cmds:
      - echo "Generating OpenAPI spec..."
      - swag init -g main.go -o ./docs --parseDependency --parseInternal
      - echo "OpenAPI spec generated at docs/swagger.yaml"

  openapi:validate:
    desc: Validate OpenAPI specification
    cmds:
      - docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli validate -i /local/docs/swagger.yaml

  openapi:serve:
    desc: Serve OpenAPI UI locally
    cmds:
      - docker run -p 8080:8080 -e SWAGGER_JSON=/docs/swagger.yaml -v ${PWD}/docs:/docs swaggerapi/swagger-ui
```

Then use:
```bash
task openapi:generate
task openapi:serve  # View at http://localhost:8080
```

---

## Quick Decision Guide

**Choose Swag if:**
- ✅ You want quick results
- ✅ You're okay with code annotations
- ✅ You're using Echo/Gin/Fiber
- ✅ You want automatic Swagger UI

**Choose go-swagger if:**
- ✅ You need advanced code generation
- ✅ You want strong validation
- ✅ You prefer different annotation style

**Choose Manual if:**
- ✅ You have a small API (< 20 endpoints)
- ✅ You want complete control
- ✅ You're importing from another tool

**Choose oapi-codegen if:**
- ✅ You're starting a new project
- ✅ You want spec-first development
- ✅ You need type-safe code generation

---

## Next Steps

1. **Install swag**: `go install github.com/swaggo/swag/cmd/swag@latest`
2. **Add basic annotations** to `main.go` (title, version, host)
3. **Annotate 2-3 key endpoints** as examples
4. **Generate spec**: `swag init -g main.go -o ./docs`
5. **Review output**: `cat docs/swagger.yaml`
6. **Add to Taskfile**: Create `openapi:generate` task
7. **Commit to git**: Include in version control

## Resources

- [Swag Documentation](https://github.com/swaggo/swag)
- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [Echo + Swag Example](https://github.com/swaggo/swag/tree/master/example/echo)
- [Go Swagger](https://goswagger.io/)
- [oapi-codegen](https://github.com/deepmap/oapi-codegen)

## Current godocs API Endpoints

All endpoints to document:

### Documents
- `GET /api/documents/latest` - Get recent documents
- `GET /api/documents/filesystem` - Get file tree
- `GET /api/document/:id` - Get document by ID
- `DELETE /api/document/*` - Delete document
- `PATCH /api/document/move/*` - Move document
- `POST /api/document/upload` - Upload document

### Folders
- `GET /api/folder/:folder` - Get folder contents
- `POST /api/folder/*` - Create folder

### Search
- `GET /api/search` - Search documents
- `POST /api/search/reindex` - Reindex search

### Admin
- `POST /api/ingest` - Trigger ingestion
- `POST /api/clean` - Clean database
- `GET /api/about` - System information

### Word Cloud
- `GET /api/wordcloud` - Get word cloud data
- `POST /api/wordcloud/recalculate` - Recalculate word cloud

### Health
- `GET /api/health` - Health check (backend only)

---

## Summary

**For godocs, I recommend Swag** because:
1. It's the most popular Go OpenAPI tool
2. Works perfectly with Echo
3. Easy to add incrementally
4. Generates Swagger UI automatically
5. Active community and good docs

Start small (annotate 5-10 endpoints) and expand as needed!
