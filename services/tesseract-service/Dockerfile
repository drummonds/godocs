FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o tesseract-service .

# Final stage
FROM alpine:latest

# Install Tesseract OCR
RUN apk add --no-cache tesseract-ocr

# Create non-root user
RUN addgroup -g 1000 tesseract && \
    adduser -D -u 1000 -G tesseract tesseract

# Copy binary from builder
COPY --from=builder /build/tesseract-service /usr/local/bin/tesseract-service

# Create temp directory with proper permissions
RUN mkdir -p /tmp && chmod 1777 /tmp

USER tesseract

EXPOSE 8001

ENTRYPOINT ["/usr/local/bin/tesseract-service"]
