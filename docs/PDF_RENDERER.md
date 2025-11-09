# PDF Renderer

godocs uses a pure Go PDF rendering engine for converting PDF pages to images for OCR processing.

## PDFium Renderer (Pure Go)

The PDFium renderer uses go-pdfium with WebAssembly to provide a pure Go solution with no CGo dependencies.

**Advantages:**
- ✅ **Pure Go** - No CGo, no external C dependencies
- ✅ **Single binary deployment** - Everything embedded in one executable
- ✅ **Cross-platform** - Works on any platform Go supports
- ✅ **Safer** - WebAssembly sandboxing prevents crashes from affecting the main process
- ✅ **No external libraries** - No need to install MuPDF or other system libraries
- ✅ **Easier deployment** - Especially useful for Docker, cloud, and embedded systems
- ✅ **CGO_ENABLED=0 compatible** - Works without C compiler

**Use cases:**
- Production deployments where simplicity and reliability are priorities
- Docker containers and cloud environments
- Systems where installing native libraries is difficult
- Embedded systems and IoT devices running gokrazy
- Cross-compilation for multiple platforms

## Implementation Details

### PDFium Renderer
- **Library:** github.com/klippa-app/go-pdfium
- **Backend:** PDFium (Google Chrome's PDF library) compiled to WebAssembly
- **Runtime:** Wazero (pure Go WebAssembly runtime)
- **Licenses:** MIT, BSD, Apache 2.0
- **DPI:** 150 (optimized for OCR quality vs. performance)

## Architecture

The PDF renderer is located in `engine/pdfrenderer/`:

```
engine/pdfrenderer/
├── renderer.go           # Interface definition
└── pdfium_renderer.go    # Pure Go WebAssembly implementation
```

## Troubleshooting

### Common Issues

**Problem:** "failed to initialize PDFium WebAssembly"
- **Solution:** Ensure sufficient memory is available. PDFium WASM requires ~50MB to initialize the pool.

**Problem:** Slow rendering performance
- **Solution:** This is expected for WebAssembly. Performance is adequate for typical home/small business use (~4 seconds per document including OCR).

**Problem:** Build failures with CGO_ENABLED=0
- **Solution:** This renderer fully supports CGO_ENABLED=0. Ensure you've removed any go-fitz references.

## Performance

On typical documents:
- **Initialization:** ~200ms (one-time pool setup)
- **Per-page rendering:** ~500-800ms at 150 DPI
- **Memory overhead:** ~50-100MB for WebAssembly pool
- **Concurrent rendering:** Thread-safe via pool management

## Future Improvements

Potential enhancements to consider:

1. **Configurable DPI** - Allow runtime DPI configuration for quality vs. speed trade-offs
2. **Page selection** - Render specific pages instead of all pages
3. **Parallel rendering** - Render multiple pages concurrently using pool workers
4. **Caching** - Cache rendered images for repeated requests
5. **Lazy initialization** - Only initialize WebAssembly pool when first PDF is encountered
