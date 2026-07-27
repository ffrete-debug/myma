package main

import (
	"ark-server-commander/config"
	"ark-server-commander/database"
	"ark-server-commander/middleware"
	"ark-server-commander/routes"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/service/update"
	"ark-server-commander/utils"
	"ark-server-commander/websocket"
	"net/http"
	"os"
	signal "os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "ark-server-commander/docs" // Import generated docs package
)

// @title ARK Server Manager API
// @version 1.0
// @description ARK Server Management System API based on Gin+Gorm
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description JWT token in the format "Bearer {token}"

func main() {
	// Initialize logger
	if err := utils.InitLogger(); err != nil {
		panic("Logger initialization failed: " + err.Error())
	}
	defer utils.Sync()

	// Initialize configuration
	if err := config.InitConfig(); err != nil {
		utils.Error("Configuration initialization failed", zap.Error(err))
		utils.Info("=========================================")
		utils.Info("💡 Solution:")
		utils.Info("1. Generate a strong random key (recommended):")
		utils.Info("   openssl rand -base64 48")
		utils.Info("")
		utils.Info("2. Set the environment variable:")
		utils.Info("   export JWT_SECRET='your-generated-secret-here'")
		utils.Info("")
		utils.Info("3. Or configure in docker-compose.yml:")
		utils.Info("   environment:")
		utils.Info("     - JWT_SECRET=your-generated-secret-here")
		utils.Info("=========================================")
		utils.Fatal("Application exiting")
	}

	// Initialize database
	database.InitDB()

	// Initialize audit log
	middleware.InitAudit(database.GetDB())

	// Initialize update monitoring hub
	updateHub := websocket.NewHub()
	go updateHub.Run()

	// Set the global hub reference so services can broadcast status changes
	websocket.SetGlobalHub(updateHub)

	// Initialize update service
	updateService := update.NewUpdateService(database.GetDB(), updateHub)

	// Check Docker environment
	if err := docker_manager.CheckDockerStatus(); err != nil {
		utils.Fatal("Docker environment check failed. Ensure Docker is installed and running", zap.Error(err))
	}
	utils.Info("Docker environment check passed")

	// Get Docker manager singleton instance
	_, err := docker_manager.GetDockerManager()
	if err != nil {
		utils.Fatal("Failed to get Docker manager", zap.Error(err))
	}
	defer docker_manager.CloseDockerManager()

	// Create Gin instance
	r := gin.New() // custom middleware, no defaults

	// Trusted proxies configurable via TRUSTED_PROXIES env (comma-separated CIDRs).
	// Default: trust nothing, so ClientIP() returns the real socket address and a
	// client-supplied X-Forwarded-For cannot forge audit log IPs or evade rate limits.
	trustedProxies := splitAndTrim(os.Getenv("TRUSTED_PROXIES"))
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		utils.Fatal("Invalid TRUSTED_PROXIES configuration", zap.Error(err))
	}

	// Request ID per request
	r.Use(middleware.RequestID())

	// Logger
	r.Use(func(c *gin.Context) {
		reqID := c.GetString("request_id")
		zap.L().Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("request_id", reqID),
			zap.String("ip", c.ClientIP()),
		)
		c.Next()
	})

	// Recovery
	r.Use(gin.Recovery())

	// Request timeout (30s)
	r.Use(middleware.Timeout(30 * time.Second))

	// Security headers
	r.Use(middleware.SecureHeaders())

	// Rate limiter: 100 requests/IP/second, burst 200
	rl := middleware.NewRateLimiter(100, 200, time.Second)
	r.Use(rl.Middleware())

	// CORS allowlist configurable via CORS_ORIGIN env (comma-separated origins).
	// Default: empty, i.e. same-origin only — no Access-Control-Allow-Origin is emitted.
	allowedOrigins := splitAndTrim(os.Getenv("CORS_ORIGIN"))
	r.Use(func(c *gin.Context) {
		// Response varies per request Origin, so it must not be cached across origins
		c.Header("Vary", "Origin")

		origin := c.GetHeader("Origin")
		for _, allowed := range allowedOrigins {
			if allowed == origin && origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Tighter rate limiter for credential endpoints: 1 request/IP/5s, burst 10.
	// Applied by path because the routes themselves are registered in routes.RegisterRoutes.
	// Registered after CORS so a 429 still carries the CORS headers the browser needs.
	authRL := middleware.NewRateLimiter(1, 10, 5*time.Second)
	authLimit := authRL.Middleware()
	r.Use(func(c *gin.Context) {
		switch c.Request.URL.Path {
		case "/api/auth/login", "/api/auth/init":
			authLimit(c)
		default:
			c.Next()
		}
	})

	// Register routes
	routes.RegisterRoutes(r, updateService, updateHub)

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		utils.Info("Received shutdown signal, gracefully shutting down...")
		docker_manager.CloseDockerManager()
		os.Exit(0)
	}()

	// Add Swagger routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	baseURL := "http://localhost:" + config.ServerPort
	utils.Info("=========================================")
	utils.Info("🚀 ARK Server Manager started successfully")
	utils.Info("📍 Server address: " + baseURL)
	utils.Info("📚 API docs: " + baseURL + "/swagger/index.html")
	utils.Info("🔗 Health check: " + baseURL + "/health")
	if len(allowedOrigins) == 0 {
		utils.Info("🌐 CORS: Same-origin only (set CORS_ORIGIN to allow cross-origin requests)")
	} else {
		utils.Info("🌐 CORS: Allowed origins: " + strings.Join(allowedOrigins, ", "))
	}
	if len(trustedProxies) == 0 {
		utils.Info("🛡️ Trusted proxies: None (client IP taken from the socket address)")
	} else {
		utils.Info("🛡️ Trusted proxies: " + strings.Join(trustedProxies, ", "))
	}
	utils.Info("🐳 Docker containerized ARK server management")
	utils.Info("🔄 Docker image background check...")
	utils.Info("📋 Docker volumes and config files initialized")
	utils.Info("📋 Server status synchronized")
	utils.Info("=========================================")

	// Explicit HTTP server with timeouts so slow clients cannot hold connections open
	srv := &http.Server{
		Addr:              ":" + config.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		utils.Fatal("Server failed to start", zap.Error(err))
	}
}

// splitAndTrim splits a comma-separated environment value into a list of
// non-empty, whitespace-trimmed entries. Returns nil when the value is empty.
func splitAndTrim(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
