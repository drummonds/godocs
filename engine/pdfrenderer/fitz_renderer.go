package pdfrenderer

import (
	"fmt"
	"image"

	"github.com/gen2brain/go-fitz"
)

// FitzRenderer implements PDF rendering using go-fitz (requires CGo and MuPDF)
type FitzRenderer struct {
}

// NewFitzRenderer creates a new Fitz-based PDF renderer
func NewFitzRenderer() (*FitzRenderer, error) {
	return &FitzRenderer{}, nil
}

// RenderPDF converts all pages of a PDF file to images using go-fitz
func (r *FitzRenderer) RenderPDF(filename string) ([]image.Image, error) {
	// Open PDF document using go-fitz
	doc, err := fitz.New(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open PDF document: %w", err)
	}
	defer doc.Close()

	// Get number of pages
	numPages := doc.NumPage()

	var images []image.Image

	// Convert each page to image at default DPI
	for pageNum := 0; pageNum < numPages; pageNum++ {
		img, err := doc.Image(pageNum)
		if err != nil {
			return nil, fmt.Errorf("unable to render page %d: %w", pageNum, err)
		}
		images = append(images, img)
	}

	return images, nil
}

// Close cleans up resources (no-op for Fitz renderer as doc is closed per-render)
func (r *FitzRenderer) Close() error {
	return nil
}
