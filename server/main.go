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
	"os"
	"syscall"
	signal "os/signal"
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

	// CORS configurable via CORS_ORIGIN env (default: * for dev)
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", corsOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
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
	utils.Info("=========================================")
	utils.Info("🚀 ARK Server Manager started successfully")
	utils.Info("📍 Server address: http://localhost:8080")
	utils.Info("📚 API docs: http://localhost:8080/swagger/index.html")
	utils.Info("🔗 Health check: http://localhost:8080/health")
	utils.Info("🌐 CORS: Enabled (all origins allowed)")
	utils.Info("🐳 Docker containerized ARK server management")
	utils.Info("🔄 Docker image background check...")
	utils.Info("📋 Docker volumes and config files initialized")
	utils.Info("📋 Server status synchronized")
	utils.Info("=========================================")

	if err := r.Run(":8080"); err != nil {
		utils.Fatal("Server failed to start", zap.Error(err))
	}
}
