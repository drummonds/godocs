package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/go-fitz"
	"github.com/ledongthuc/pdf"
)

type ExtractTextRequest struct {
	PDF []byte `json:"pdf"`
}

type ExtractTextResponse struct {
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

type ToImageRequest struct {
	PDF []byte `json:"pdf"`
}

type ToImageResponse struct {
	Image string `json:"image"` // base64 encoded PNG
	Error string `json:"error,omitempty"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	log.Printf("Starting PDF service on port %s", port)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/pdf/extract-text", extractTextHandler)
	http.HandleFunc("/pdf/to-image", toImageHandler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func extractTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("pdf")
	if err != nil {
		sendErrorResponse(w, "No PDF file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Processing text extraction for file: %s", header.Filename)

	// Read file content
	pdfData, err := io.ReadAll(file)
	if err != nil {
		sendErrorResponse(w, "Failed to read PDF file", http.StatusInternalServerError)
		return
	}

	// Extract text
	text, err := extractText(pdfData)
	if err != nil {
		log.Printf("Text extraction error: %v", err)
		sendErrorResponse(w, fmt.Sprintf("Text extraction failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := ExtractTextResponse{
		Text: text,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func toImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		sendErrorResponse(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("pdf")
	if err != nil {
		sendErrorResponse(w, "No PDF file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Processing PDF to image conversion for file: %s", header.Filename)

	// Read file content
	pdfData, err := io.ReadAll(file)
	if err != nil {
		sendErrorResponse(w, "Failed to read PDF file", http.StatusInternalServerError)
		return
	}

	// Convert to image
	imageData, err := convertToImage(pdfData)
	if err != nil {
		log.Printf("Image conversion error: %v", err)
		sendErrorResponse(w, fmt.Sprintf("Image conversion failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := ToImageResponse{
		Image: imageData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func extractText(pdfData []byte) (string, error) {
	reader := bytes.NewReader(pdfData)

	pdfReader, err := pdf.NewReader(reader, int64(len(pdfData)))
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	totalPages := pdfReader.NumPage()
	var fullText string

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			log.Printf("Warning: failed to extract text from page %d: %v", pageNum, err)
			continue
		}

		fullText += text
	}

	return fullText, nil
}

func convertToImage(pdfData []byte) (string, error) {
	// Create a temporary file for the PDF
	tempFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write PDF data to temp file
	if _, err := tempFile.Write(pdfData); err != nil {
		return "", fmt.Errorf("failed to write PDF to temp file: %w", err)
	}
	tempFile.Close()

	// Open PDF with go-fitz
	doc, err := fitz.New(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	numPages := doc.NumPage()
	if numPages == 0 {
		return "", fmt.Errorf("PDF has no pages")
	}

	var images []image.Image

	// Convert each page to image
	for pageNum := 0; pageNum < numPages; pageNum++ {
		img, err := doc.Image(pageNum)
		if err != nil {
			return "", fmt.Errorf("failed to render page %d: %w", pageNum, err)
		}
		images = append(images, img)
	}

	// Combine images vertically
	var combinedImage image.Image
	if len(images) == 1 {
		combinedImage = images[0]
	} else {
		// Calculate total height
		totalHeight := 0
		maxWidth := 0
		for _, img := range images {
			bounds := img.Bounds()
			totalHeight += bounds.Dy()
			if bounds.Dx() > maxWidth {
				maxWidth = bounds.Dx()
			}
		}

		// Create new image
		combined := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))

		// Draw each page
		currentY := 0
		for _, img := range images {
			bounds := img.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					combined.Set(x, currentY+y-bounds.Min.Y, img.At(x, y))
				}
			}
			currentY += bounds.Dy()
		}

		combinedImage = combined
	}

	// Resize to 1024px width
	resized := imaging.Resize(combinedImage, 1024, 0, imaging.Lanczos)

	// Apply sharpening
	sharpened := imaging.Sharpen(resized, 0.5)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, sharpened); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	// Return base64 encoded image
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return encoded, nil
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{
		"error": message,
	}
	json.NewEncoder(w).Encode(response)
}
