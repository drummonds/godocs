# goEDMS Microservices Architecture

This document describes the microservices architecture for goEDMS, which separates heavy dependencies (Tesseract OCR and PDF rendering) into standalone services.

## Architecture Overview

goEDMS now consists of three separate services:

1. **Main Application** (Pure Go) - Document management, web UI, database
2. **Tesseract OCR Service** - Optical character recognition for images
3. **PDF Service** - PDF text extraction and PDF-to-image conversion

### Benefits

- **Pure Go Main Application**: No CGO dependencies, easier deployment
- **Scalability**: Services can be scaled independently
- **Isolation**: Heavy processing isolated from main application
- **Maintainability**: Each service has a single responsibility
- **Flexibility**: Easy to swap implementations or add new services

## Services

### 1. Main Application (goEDMS)

**Port**: 8000
**Language**: Pure Go (no CGO)
**Responsibilities**:
- Web UI and API
- Document management
- Database operations
- User authentication
- Scheduling and ingress processing

**Configuration**:
- `TESSERACT_SERVICE_URL`: URL of Tesseract service (default: `http://tesseract-service:8001`)
- `PDF_SERVICE_URL`: URL of PDF service (default: `http://pdf-service:8002`)

### 2. Tesseract OCR Service

**Port**: 8001
**Language**: Go with Tesseract binary
**Responsibilities**:
- Optical character recognition
- Image text extraction

**API Endpoints**:
- `GET /health` - Health check
- `POST /ocr` - Extract text from image

**Example Usage**:
```bash
# Health check
curl http://localhost:8001/health

# OCR an image
curl -X POST -F "image=@document.png" http://localhost:8001/ocr
```

### 3. PDF Service

**Port**: 8002
**Language**: Go with MuPDF (via go-fitz)
**Responsibilities**:
- PDF text extraction
- PDF to image conversion

**API Endpoints**:
- `GET /health` - Health check
- `POST /pdf/extract-text` - Extract text from PDF
- `POST /pdf/to-image` - Convert PDF to PNG image

**Example Usage**:
```bash
# Health check
curl http://localhost:8002/health

# Extract text from PDF
curl -X POST -F "pdf=@document.pdf" http://localhost:8002/pdf/extract-text

# Convert PDF to image
curl -X POST -F "pdf=@document.pdf" http://localhost:8002/pdf/to-image
```

## Deployment

### Docker Compose (Recommended)

```bash
# Build and start all services
docker-compose -f docker-compose.microservices.yml up --build

# Start in detached mode
docker-compose -f docker-compose.microservices.yml up -d

# View logs
docker-compose -f docker-compose.microservices.yml logs -f

# Stop all services
docker-compose -f docker-compose.microservices.yml down
```

### Individual Services

Each service can be built and run independently:

```bash
# Tesseract Service
cd services/tesseract-service
docker build -t tesseract-service .
docker run -p 8001:8001 tesseract-service

# PDF Service
cd services/pdf-service
docker build -t pdf-service .
docker run -p 8002:8002 pdf-service

# Main Application
docker build -t goedms .
docker run -p 8000:8000 \
  -e TESSERACT_SERVICE_URL=http://localhost:8001 \
  -e PDF_SERVICE_URL=http://localhost:8002 \
  goedms
```

### Local Development

For local development without Docker:

```bash
# Terminal 1: Start Tesseract service
cd services/tesseract-service
go run main.go

# Terminal 2: Start PDF service
cd services/pdf-service
go run main.go

# Terminal 3: Start main application
export TESSERACT_SERVICE_URL=http://localhost:8001
export PDF_SERVICE_URL=http://localhost:8002
go run main.go
```

## Configuration

### Environment Variables

Copy `.env.microservices.example` to `.env` and customize:

```bash
cp .env.microservices.example .env
```

Key configuration options:

| Variable | Description | Default |
|----------|-------------|---------|
| `TESSERACT_SERVICE_URL` | Tesseract OCR service URL | `http://tesseract-service:8001` |
| `PDF_SERVICE_URL` | PDF processing service URL | `http://pdf-service:8002` |
| `DATABASE_TYPE` | Database type (sqlite/postgres) | `sqlite` |
| `DOCUMENT_PATH` | Document storage path | `/documents` |
| `INGRESS_PATH` | Ingress folder path | `/ingress` |

## Migration from Monolithic Version

To migrate from the previous monolithic version:

1. **Data Migration**: Your existing data and documents are compatible
2. **Configuration**: Update environment variables to include service URLs
3. **Deployment**: Use new Docker Compose file or deploy services separately

No database migrations are required.

## Monitoring

### Health Checks

Each service provides a `/health` endpoint:

```bash
# Check all services
curl http://localhost:8000/about  # Main application
curl http://localhost:8001/health # Tesseract service
curl http://localhost:8002/health # PDF service
```

### Service Status

The Docker Compose configuration includes health checks that automatically monitor service health.

### Logs

```bash
# View all logs
docker-compose -f docker-compose.microservices.yml logs -f

# View specific service logs
docker-compose -f docker-compose.microservices.yml logs -f goedms
docker-compose -f docker-compose.microservices.yml logs -f tesseract-service
docker-compose -f docker-compose.microservices.yml logs -f pdf-service
```

## Troubleshooting

### Service Connection Issues

If the main application cannot connect to services:

1. Check service health: `curl http://localhost:8001/health`
2. Verify network connectivity between containers
3. Check Docker network: `docker network inspect goedms-network`
4. Review logs: `docker-compose logs tesseract-service pdf-service`

### OCR Not Working

1. Verify Tesseract service is running: `curl http://localhost:8001/health`
2. Check service logs: `docker-compose logs tesseract-service`
3. Test service directly: `curl -X POST -F "image=@test.png" http://localhost:8001/ocr`

### PDF Processing Issues

1. Verify PDF service is running: `curl http://localhost:8002/health`
2. Check service logs: `docker-compose logs pdf-service`
3. Test service directly: `curl -X POST -F "pdf=@test.pdf" http://localhost:8002/pdf/extract-text`

## Performance Tuning

### Scaling Services

Services can be scaled independently using Docker Compose:

```bash
# Scale Tesseract service to 3 instances
docker-compose -f docker-compose.microservices.yml up --scale tesseract-service=3

# Scale PDF service to 2 instances
docker-compose -f docker-compose.microservices.yml up --scale pdf-service=2
```

### Resource Limits

Add resource limits in `docker-compose.microservices.yml`:

```yaml
services:
  tesseract-service:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

## Security Considerations

1. **Network Isolation**: Services communicate over internal Docker network
2. **No Direct External Access**: Only main application exposed (port 8000)
3. **Service Authentication**: Consider adding API keys for production
4. **Resource Limits**: Set memory/CPU limits to prevent DoS

## Development

### Adding New Services

To add a new service:

1. Create service directory: `services/new-service/`
2. Implement HTTP API with `/health` endpoint
3. Add to Docker Compose configuration
4. Update main application to call new service
5. Document in this file

### Testing

Each service includes a README with testing instructions. See:
- `services/tesseract-service/README.md`
- `services/pdf-service/README.md`

## Future Enhancements

- Load balancing for multiple service instances
- Service discovery (Consul, etcd)
- API gateway (Kong, Traefik)
- Message queue for asynchronous processing (RabbitMQ, Redis)
- Metrics and monitoring (Prometheus, Grafana)
- Distributed tracing (Jaeger, Zipkin)
