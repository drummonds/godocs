package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	config "github.com/drummonds/godocs/config"
	"github.com/drummonds/godocs/webapp"
)

// Logger is global since we will need it everywhere
var Logger *slog.Logger

func main() {
	// Parse command-line flags
	port := flag.String("port", "3000", "Port to run frontend server on")
	apiURL := flag.String("api", "", "Backend API URL (overrides config)")
	flag.Parse()

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎨  godocs Frontend Server")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("• WASM application server")
	fmt.Println("• Proxies API calls to backend")
	fmt.Println(strings.Repeat("=", 50) + "\n")

	frontendConfig, logger := config.SetupFrontend()
	Logger = logger
	config.Logger = logger

	// Override API URL if provided via flag
	if *apiURL != "" {
		frontendConfig.ServerAPIURL = *apiURL
	}

	Logger.Info("Frontend server starting",
		"backendAPI", frontendConfig.ServerAPIURL,
		"port", *port)

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// CORS - allow requests from anywhere (since we're just serving static content)
	e.Use(middleware.CORS())

	// Request logging
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))

	// Serve the go-app WASM handler
	Logger.Info("Setting up WASM application...")
	appHandler := webapp.Handler()

	// Serve wasm_exec.js
	e.GET("/wasm_exec.js", func(c echo.Context) error {
		return c.File("web/wasm_exec.js")
	})

	// Register go-app specific resources
	e.GET("/app.js", echo.WrapHandler(appHandler))
	e.GET("/app.css", echo.WrapHandler(appHandler))
	e.GET("/manifest.webmanifest", echo.WrapHandler(appHandler))

	// Serve static assets
	e.Static("/web", "web")
	e.File("/webapp/webapp.css", "webapp/webapp.css")
	e.File("/webapp/wordcloud.css", "webapp/wordcloud.css")
	e.File("/favicon.ico", "public/built/favicon.ico")

	// Inject backend API URL into the page
	// This will be available as a global JavaScript variable
	e.GET("/config.js", func(c echo.Context) error {
		configJS := fmt.Sprintf(`
// godocs Frontend Configuration
window.godocsConfig = {
    apiURL: "%s",
    newDocumentCount: %d
};
console.log("godocs Config loaded:", window.godocsConfig);
`, frontendConfig.ServerAPIURL, frontendConfig.NewDocumentNumber)
		c.Response().Header().Set("Content-Type", "application/javascript")
		return c.String(http.StatusOK, configJS)
	})

	// API proxy middleware - forward /api/* requests to backend
	backendURL := mustParseURL(frontendConfig.ServerAPIURL)
	e.Group("/api", middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
			{
				URL: backendURL,
			},
		}),
	}))

	// Serve go-app handler for all other routes (must be last)
	e.Any("/*", echo.WrapHandler(appHandler))

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	Logger.Info("Starting Frontend Server", "address", addr, "backendAPI", frontendConfig.ServerAPIURL)
	fmt.Printf("\n✅  Frontend Server running on %s\n", addr)
	fmt.Printf("🎨  Open http://localhost:%s in your browser\n", *port)
	fmt.Printf("📡  API proxied to: %s\n\n", frontendConfig.ServerAPIURL)

	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		Logger.Error("Server failed to start", "error", err)
	}
}

// mustParseURL parses a URL and panics if invalid
func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("Invalid URL: %s - %v", rawURL, err))
	}
	return u
}
