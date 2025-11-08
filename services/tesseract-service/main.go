package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type OCRRequest struct {
	Image []byte `json:"image"`
}

type OCRResponse struct {
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Tesseract string `json:"tesseract"`
	Timestamp string `json:"timestamp"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	tesseractPath := os.Getenv("TESSERACT_PATH")
	if tesseractPath == "" {
		tesseractPath = "/usr/bin/tesseract"
	}

	// Verify Tesseract is available
	if _, err := os.Stat(tesseractPath); os.IsNotExist(err) {
		log.Fatalf("Tesseract not found at %s", tesseractPath)
	}

	log.Printf("Starting Tesseract OCR service on port %s", port)
	log.Printf("Using Tesseract at: %s", tesseractPath)

	http.HandleFunc("/health", healthHandler(tesseractPath))
	http.HandleFunc("/ocr", ocrHandler(tesseractPath))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(tesseractPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check Tesseract version
		cmd := exec.Command(tesseractPath, "--version")
		output, err := cmd.CombinedOutput()
		tesseractInfo := "available"
		if err != nil {
			tesseractInfo = fmt.Sprintf("error: %v", err)
		} else {
			tesseractInfo = string(bytes.Split(output, []byte("\n"))[0])
		}

		response := HealthResponse{
			Status:    "healthy",
			Tesseract: tesseractInfo,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func ocrHandler(tesseractPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		file, header, err := r.FormFile("image")
		if err != nil {
			sendErrorResponse(w, "No image file provided", http.StatusBadRequest)
			return
		}
		defer file.Close()

		log.Printf("Processing OCR request for file: %s", header.Filename)

		// Read file content
		imageData, err := io.ReadAll(file)
		if err != nil {
			sendErrorResponse(w, "Failed to read image file", http.StatusInternalServerError)
			return
		}

		// Process OCR
		text, err := processOCR(tesseractPath, imageData, header.Filename)
		if err != nil {
			log.Printf("OCR processing error: %v", err)
			sendErrorResponse(w, fmt.Sprintf("OCR processing failed: %v", err), http.StatusInternalServerError)
			return
		}

		response := OCRResponse{
			Text: text,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func processOCR(tesseractPath string, imageData []byte, filename string) (string, error) {
	// Create a temporary directory for this request
	tempDir, err := os.MkdirTemp("", "ocr-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Determine file extension
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png"
	}

	// Write image to temporary file
	inputPath := filepath.Join(tempDir, "input"+ext)
	if err := os.WriteFile(inputPath, imageData, 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// Output path (without extension - Tesseract adds .txt)
	outputBase := filepath.Join(tempDir, "output")
	outputPath := outputBase + ".txt"

	// Run Tesseract
	cmd := exec.Command(tesseractPath, inputPath, outputBase)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tesseract command failed: %w, stderr: %s", err, stderr.String())
	}

	// Read the output
	textData, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read OCR output: %w", err)
	}

	return string(textData), nil
}

func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := OCRResponse{
		Error: message,
	}
	json.NewEncoder(w).Encode(response)
}
