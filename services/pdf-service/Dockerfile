FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies for go-fitz (requires MuPDF)
RUN apk add --no-cache \
    gcc \
    g++ \
    musl-dev \
    mupdf-dev

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o pdf-service .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    mupdf \
    mupdf-tools

# Create non-root user
RUN addgroup -g 1000 pdfservice && \
    adduser -D -u 1000 -G pdfservice pdfservice

# Copy binary from builder
COPY --from=builder /build/pdf-service /usr/local/bin/pdf-service

# Create temp directory with proper permissions
RUN mkdir -p /tmp && chmod 1777 /tmp

USER pdfservice

EXPOSE 8002

ENTRYPOINT ["/usr/local/bin/pdf-service"]
