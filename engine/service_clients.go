package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ServiceClients holds HTTP clients for external services
type ServiceClients struct {
	TesseractURL string
	PDFURL       string
	HTTPClient   *http.Client
}

// NewServiceClients creates a new service client manager
func NewServiceClients(tesseractURL, pdfURL string) *ServiceClients {
	return &ServiceClients{
		TesseractURL: tesseractURL,
		PDFURL:       pdfURL,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// OCRResponse represents the response from the Tesseract service
type OCRResponse struct {
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

// PDFTextResponse represents the response from PDF text extraction
type PDFTextResponse struct {
	Text  string `json:"text"`
	Error string `json:"error,omitempty"`
}

// PDFImageResponse represents the response from PDF to image conversion
type PDFImageResponse struct {
	Image string `json:"image"` // base64 encoded PNG
	Error string `json:"error,omitempty"`
}

// CallOCRService sends an image to the Tesseract service and returns extracted text
func (sc *ServiceClients) CallOCRService(imagePath string) (string, error) {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Make HTTP request
	url := fmt.Sprintf("%s/ocr", sc.TesseractURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := sc.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OCR service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OCR service returned error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var ocrResp OCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
		return "", fmt.Errorf("failed to decode OCR response: %w", err)
	}

	if ocrResp.Error != "" {
		return "", fmt.Errorf("OCR service error: %s", ocrResp.Error)
	}

	return ocrResp.Text, nil
}

// CallPDFExtractText sends a PDF to the PDF service and returns extracted text
func (sc *ServiceClients) CallPDFExtractText(pdfPath string) (string, error) {
	// Open the PDF file
	file, err := os.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("pdf", filepath.Base(pdfPath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Make HTTP request
	url := fmt.Sprintf("%s/pdf/extract-text", sc.PDFURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := sc.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call PDF service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("PDF service returned error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var pdfResp PDFTextResponse
	if err := json.NewDecoder(resp.Body).Decode(&pdfResp); err != nil {
		return "", fmt.Errorf("failed to decode PDF response: %w", err)
	}

	if pdfResp.Error != "" {
		return "", fmt.Errorf("PDF service error: %s", pdfResp.Error)
	}

	return pdfResp.Text, nil
}

// CallPDFToImage sends a PDF to the PDF service and returns a PNG image (saved to disk)
func (sc *ServiceClients) CallPDFToImage(pdfPath string, outputImagePath string) error {
	// Open the PDF file
	file, err := os.Open(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("pdf", filepath.Base(pdfPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Make HTTP request
	url := fmt.Sprintf("%s/pdf/to-image", sc.PDFURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := sc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call PDF service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("PDF service returned error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var pdfResp PDFImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&pdfResp); err != nil {
		return fmt.Errorf("failed to decode PDF response: %w", err)
	}

	if pdfResp.Error != "" {
		return fmt.Errorf("PDF service error: %s", pdfResp.Error)
	}

	// Decode base64 image
	imageData, err := base64.StdEncoding.DecodeString(pdfResp.Image)
	if err != nil {
		return fmt.Errorf("failed to decode base64 image: %w", err)
	}

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputImagePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write image to disk
	if err := os.WriteFile(outputImagePath, imageData, 0644); err != nil {
		return fmt.Errorf("failed to write image file: %w", err)
	}

	return nil
}
