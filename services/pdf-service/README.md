# PDF Service

A lightweight HTTP service that provides PDF processing capabilities including text extraction and PDF-to-image conversion.

## API Endpoints

### Health Check
```bash
GET /health
```

Returns service health status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-08T12:00:00Z"
}
```

### Extract Text from PDF
```bash
POST /pdf/extract-text
Content-Type: multipart/form-data
```

Extracts text content from a PDF file.

**Request:**
- `pdf` (file): PDF file

**Response:**
```json
{
  "text": "Extracted text from PDF..."
}
```

### Convert PDF to Image
```bash
POST /pdf/to-image
Content-Type: multipart/form-data
```

Converts PDF pages to a single PNG image (pages combined vertically). Returns base64-encoded PNG.

**Request:**
- `pdf` (file): PDF file

**Response:**
```json
{
  "image": "iVBORw0KGgoAAAANSUhEUgAA... (base64 encoded PNG)"
}
```

**Error Response:**
```json
{
  "error": "Error message"
}
```

## Environment Variables

- `PORT` - HTTP server port (default: 8002)

## Building

### Local Build
```bash
# Requires MuPDF development libraries
go build -o pdf-service
```

### Docker Build
```bash
docker build -t pdf-service .
```

## Running

### Local
```bash
./pdf-service
```

### Docker
```bash
docker run -p 8002:8002 pdf-service
```

## Testing

```bash
# Health check
curl http://localhost:8002/health

# Extract text
curl -X POST -F "pdf=@document.pdf" http://localhost:8002/pdf/extract-text

# Convert to image
curl -X POST -F "pdf=@document.pdf" http://localhost:8002/pdf/to-image
```

## Dependencies

This service uses:
- **go-fitz**: MuPDF bindings for Go (PDF rendering)
- **ledongthuc/pdf**: Pure Go PDF text extraction
- **disintegration/imaging**: Image processing

Note: go-fitz requires CGO and MuPDF libraries, which is why this service is containerized separately.
