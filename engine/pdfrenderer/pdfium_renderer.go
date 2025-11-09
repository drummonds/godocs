package pdfrenderer

import (
	"fmt"
	"image"
	"os"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
)

// PDFiumRenderer implements PDF rendering using go-pdfium with WebAssembly (pure Go, no CGo)
type PDFiumRenderer struct {
	pool     pdfium.Pool
	instance pdfium.Pdfium
}

// NewPDFiumRenderer creates a new PDFium-based PDF renderer using WebAssembly
func NewPDFiumRenderer() (*PDFiumRenderer, error) {
	// Initialize WebAssembly pool with minimal configuration
	// For single-threaded usage, we keep it simple
	pool, err := webassembly.Init(webassembly.Config{
		MinIdle:  1, // Minimum idle workers
		MaxIdle:  1, // Maximum idle workers
		MaxTotal: 1, // Total worker limit
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PDFium WebAssembly: %w", err)
	}

	// Get a PDFium instance from the pool
	instance, err := pool.GetInstance(time.Second * 30)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to get PDFium instance: %w", err)
	}

	return &PDFiumRenderer{
		pool:     pool,
		instance: instance,
	}, nil
}

// RenderPDF converts all pages of a PDF file to images using go-pdfium WebAssembly
func (r *PDFiumRenderer) RenderPDF(filename string) ([]image.Image, error) {
	// Read the PDF file
	pdfBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read PDF file: %w", err)
	}

	// Open the PDF document
	doc, err := r.instance.OpenDocument(&requests.OpenDocument{
		File: &pdfBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to open PDF document: %w", err)
	}
	defer r.instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{
		Document: doc.Document,
	})

	// Get the number of pages
	pageCountResp, err := r.instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{
		Document: doc.Document,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get page count: %w", err)
	}

	numPages := pageCountResp.PageCount
	images := make([]image.Image, 0, numPages)

	// Render each page at 150 DPI (optimized for OCR quality)
	for pageIndex := 0; pageIndex < numPages; pageIndex++ {
		pageRender, err := r.instance.RenderPageInDPI(&requests.RenderPageInDPI{
			DPI: 150, // Match the DPI mentioned in original convertToImage function
			Page: requests.Page{
				ByIndex: &requests.PageByIndex{
					Document: doc.Document,
					Index:    pageIndex,
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("unable to render page %d: %w", pageIndex, err)
		}

		// Extract the image from the result
		images = append(images, pageRender.Result.Image)

		// Clean up WebAssembly resources for this page
		pageRender.Cleanup()
	}

	return images, nil
}

// Close cleans up resources used by the PDFium renderer
func (r *PDFiumRenderer) Close() error {
	if r.pool != nil {
		r.pool.Close()
		r.pool = nil
	}
	r.instance = nil
	return nil
}
