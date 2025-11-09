package pdfrenderer

import (
	"image"
)

// Renderer defines the interface for PDF to image conversion
type Renderer interface {
	// RenderPDF converts all pages of a PDF file to images
	// Returns a slice of images, one per page
	RenderPDF(filename string) ([]image.Image, error)

	// Close cleans up any resources used by the renderer
	Close() error
}

// NewRenderer creates a new PDFium-based PDF renderer (pure Go, no CGo)
func NewRenderer() (Renderer, error) {
	return NewPDFiumRenderer()
}
