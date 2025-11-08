# PDF Renderer Options

goEDMS supports two different PDF rendering engines for converting PDF pages to images for OCR processing.

## Available Renderers

### 1. PDFium (Pure Go) - **RECOMMENDED**

**Type:** `pdfium` (default)

The PDFium renderer uses go-pdfium with WebAssembly to provide a pure Go solution with no CGo dependencies.

**Advantages:**
- ✅ **Pure Go** - No CGo, no external C dependencies
- ✅ **Single binary deployment** - Embed everything in one executable
- ✅ **Cross-platform** - Works on any platform Go supports
- ✅ **Safer** - WebAssembly sandboxing prevents crashes from affecting the main process
- ✅ **No external libraries** - No need to install MuPDF or other system libraries
- ✅ **Easier deployment** - Especially useful for Docker, cloud, and embedded systems

**Trade-offs:**
- ⚠️ About 2x slower than CGo-based rendering
- ⚠️ Higher memory usage due to WebAssembly overhead

**Use cases:**
- Production deployments where simplicity and reliability are priorities
- Docker containers and cloud environments
- Systems where installing native libraries is difficult
- Embedded systems and IoT devices running gokrazy

### 2. Fitz (CGo-based)

**Type:** `fitz`

The Fitz renderer uses go-fitz which wraps the MuPDF C library via CGo.

**Advantages:**
- ✅ Faster rendering (approximately 2x faster than PDFium WASM)
- ✅ Lower memory usage
- ✅ Mature and battle-tested library

**Trade-offs:**
- ⚠️ **Requires CGo** - Must have a C compiler available
- ⚠️ **External dependency** - Requires MuPDF library to be installed
- ⚠️ **Complex deployment** - More difficult to package and deploy
- ⚠️ **Platform-specific** - Must compile for each target platform
- ⚠️ **Crash risk** - C code can crash the entire Go process

**Use cases:**
- High-volume document processing where performance is critical
- Environments where native libraries are easy to manage
- Systems with existing MuPDF installations

## Configuration

Set the PDF renderer using the `PDF_RENDERER` environment variable:

```bash
# Use PDFium (pure Go, default)
PDF_RENDERER=pdfium

# Use Fitz (CGo-based)
PDF_RENDERER=fitz
```

### Example `.env` file:

```env
# PDF Rendering Configuration
# Options: "pdfium" (pure Go, default) or "fitz" (CGo-based)
PDF_RENDERER=pdfium
```

## Performance Comparison

Based on typical document workloads:

| Renderer | Speed | Memory | Deployment | Recommendation |
|----------|-------|--------|------------|----------------|
| PDFium   | ~2x slower | Higher | ⭐ Simple | **Production default** |
| Fitz     | Faster | Lower | Complex | High-volume processing |

## Implementation Details

### PDFium Renderer
- **Library:** github.com/klippa-app/go-pdfium
- **Backend:** PDFium (Google Chrome's PDF library) compiled to WebAssembly
- **Runtime:** Wazero (pure Go WebAssembly runtime)
- **Licenses:** MIT, BSD, Apache 2.0
- **DPI:** 150 (matches Fitz implementation)

### Fitz Renderer
- **Library:** github.com/gen2brain/go-fitz
- **Backend:** MuPDF C library
- **Runtime:** CGo
- **Licenses:** AGPL (MuPDF), Proprietary (commercial)
- **DPI:** Native rendering

## Architecture

The PDF renderer is abstracted through a clean interface in `engine/pdfrenderer/`:

```
engine/pdfrenderer/
├── renderer.go           # Interface definition and factory
├── pdfium_renderer.go    # Pure Go implementation
└── fitz_renderer.go      # CGo implementation
```

The engine automatically selects the configured renderer at runtime without requiring code changes.

## Troubleshooting

### PDFium Issues

**Problem:** "failed to initialize PDFium WebAssembly"
- **Solution:** Ensure sufficient memory is available. PDFium WASM requires memory to initialize.

**Problem:** Slow rendering performance
- **Solution:** This is expected (~2x slower than CGo). Consider using Fitz for high-volume workloads.

### Fitz Issues

**Problem:** "Unable to open PDF document" with Fitz
- **Solution:** Ensure MuPDF is installed: `apt-get install libmupdf-dev` (Debian/Ubuntu)

**Problem:** Build errors with CGo
- **Solution:** Install a C compiler: `apt-get install build-essential`

**Problem:** Cross-compilation issues
- **Solution:** Switch to PDFium renderer for simpler cross-platform builds

## Migration Guide

### Switching from Fitz to PDFium

1. Set `PDF_RENDERER=pdfium` in your environment
2. Remove MuPDF system dependencies (optional)
3. Rebuild your application (no code changes needed)
4. Test with your typical document workload
5. Monitor performance and memory usage

### Switching from PDFium to Fitz

1. Install MuPDF: `apt-get install libmupdf-dev`
2. Ensure CGo is enabled: `CGO_ENABLED=1`
3. Set `PDF_RENDERER=fitz` in your environment
4. Rebuild your application
5. Test with your typical document workload

## Recommendations

- **Default:** Use PDFium for ease of deployment and reliability
- **High-volume:** Consider Fitz if processing thousands of documents daily
- **Docker/Cloud:** Use PDFium for simpler container images
- **Embedded (gokrazy):** Use PDFium for pure Go deployment
- **Development:** Use PDFium for easier setup

## Future Improvements

Potential enhancements to consider:

1. **Configurable DPI** - Allow runtime DPI configuration
2. **Page selection** - Render specific pages instead of all pages
3. **Parallel rendering** - Render multiple pages concurrently
4. **Caching** - Cache rendered images for repeated requests
5. **Alternative backends** - Support for additional PDF libraries
