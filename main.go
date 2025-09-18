package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"url-shortener/config"
	"url-shortener/database"
	"url-shortener/handlers"
	"url-shortener/middleware"
	"url-shortener/models"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize services
	urlService := models.NewURLService(db)
	urlHandler := handlers.NewURLHandler(urlService, cfg.BaseURL)

	// Set up Gin router
	gin.SetMode(gin.ReleaseMode) // Set to gin.DebugMode for development
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimiter())
	router.Use(gin.Recovery())

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/urls", urlHandler.CreateShortURL)
		api.GET("/urls", urlHandler.ListURLs)
		api.GET("/urls/:shortCode/stats", urlHandler.GetURLStats)
		api.DELETE("/urls/:shortCode", urlHandler.DeleteURL)
	}

	// Health check
	router.GET("/health", urlHandler.HealthCheck)

	// Redirect route (should be last to avoid conflicts)
	router.GET("/:shortCode", urlHandler.RedirectURL)

	// Serve a simple HTML page for the root path
	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>URL Shortener</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .api-section {
            margin: 20px 0;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 5px;
            border-left: 4px solid #007bff;
        }
        .endpoint {
            font-family: monospace;
            background: #e9ecef;
            padding: 10px;
            border-radius: 3px;
            margin: 10px 0;
        }
        .method {
            font-weight: bold;
            color: #007bff;
        }
        pre {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔗 URL Shortener Service</h1>
        <p>Welcome to the URL Shortener API! This service allows you to create short links and track their usage.</p>
        
        <div class="api-section">
            <h3>API Endpoints</h3>
            
            <h4>Create Short URL</h4>
            <div class="endpoint">
                <span class="method">POST</span> /api/v1/urls
            </div>
            <pre>{
  "original_url": "https://example.com/very/long/url",
  "custom_code": "my-link" (optional),
  "expires_in": 24 (optional, hours)
}</pre>

            <h4>Get URL Statistics</h4>
            <div class="endpoint">
                <span class="method">GET</span> /api/v1/urls/{shortCode}/stats
            </div>

            <h4>List All URLs</h4>
            <div class="endpoint">
                <span class="method">GET</span> /api/v1/urls?page=1&per_page=10
            </div>

            <h4>Delete URL</h4>
            <div class="endpoint">
                <span class="method">DELETE</span> /api/v1/urls/{shortCode}
            </div>

            <h4>Redirect (Short URL)</h4>
            <div class="endpoint">
                <span class="method">GET</span> /{shortCode}
            </div>

            <h4>Health Check</h4>
            <div class="endpoint">
                <span class="method">GET</span> /health
            </div>
        </div>

        <div class="api-section">
            <h3>Example Usage</h3>
            <pre>
# Create a short URL
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"original_url": "https://example.com"}'

# Access short URL (redirects)
curl -L http://localhost:8080/abc123

# Get statistics
curl http://localhost:8080/api/v1/urls/abc123/stats
            </pre>
        </div>
    </div>
</body>
</html>
		`))
	})

	// Start server
	log.Printf("Starting URL Shortener server on port %s", cfg.ServerPort)
	log.Printf("Base URL: %s", cfg.BaseURL)
	
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}