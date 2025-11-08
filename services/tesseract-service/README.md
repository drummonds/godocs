# Tesseract OCR Service

A lightweight HTTP service that provides OCR (Optical Character Recognition) capabilities using Tesseract.

## API Endpoints

### Health Check
```bash
GET /health
```

Returns service health status and Tesseract version information.

**Response:**
```json
{
  "status": "healthy",
  "tesseract": "tesseract 5.x.x",
  "timestamp": "2025-11-08T12:00:00Z"
}
```

### OCR Processing
```bash
POST /ocr
Content-Type: multipart/form-data
```

Processes an image file and returns extracted text.

**Request:**
- `image` (file): Image file (TIFF, JPG, JPEG, PNG)

**Response:**
```json
{
  "text": "Extracted text from the image..."
}
```

**Error Response:**
```json
{
  "error": "Error message"
}
```

## Environment Variables

- `PORT` - HTTP server port (default: 8001)
- `TESSERACT_PATH` - Path to Tesseract binary (default: /usr/bin/tesseract)

## Building

### Local Build
```bash
go build -o tesseract-service
```

### Docker Build
```bash
docker build -t tesseract-service .
```

## Running

### Local
```bash
./tesseract-service
```

### Docker
```bash
docker run -p 8001:8001 tesseract-service
```

## Testing

```bash
# Health check
curl http://localhost:8001/health

# OCR processing
curl -X POST -F "image=@test.png" http://localhost:8001/ocr
```
