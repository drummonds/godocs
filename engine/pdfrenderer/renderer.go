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

// RendererType specifies which PDF rendering implementation to use
type RendererType string

const (
	// RendererTypeFitz uses go-fitz (CGo, requires MuPDF)
	RendererTypeFitz RendererType = "fitz"

	// RendererTypePDFium uses go-pdfium with WebAssembly (pure Go, no CGo)
	RendererTypePDFium RendererType = "pdfium"
)

// NewRenderer creates a new PDF renderer of the specified type
func NewRenderer(rendererType RendererType) (Renderer, error) {
	switch rendererType {
	case RendererTypeFitz:
		return NewFitzRenderer()
	case RendererTypePDFium:
		return NewPDFiumRenderer()
	default:
		return NewPDFiumRenderer() // Default to pure Go implementation
	}
}
