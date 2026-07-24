package routes

import (
	"ark-server-commander/controllers/auth"
	"ark-server-commander/controllers/images"
	"ark-server-commander/controllers/plugins"
	"ark-server-commander/controllers/servers"
	"ark-server-commander/middleware"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"ark-server-commander/service/update"
	"ark-server-commander/websocket"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, updateService *update.UpdateService, hub *websocket.Hub) {
	// （）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Servers "})
	})

	//  - 
	// YesNo
	if _, err := os.Stat("./static"); err == nil {
		//  Next.js 
		r.Static("/_next", "./static/_next")
		r.Static("/public", "./static/public")
		r.StaticFile("/favicon.ico", "./static/public/favicon.ico")

		//  SPA  -  API  index.html
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// Yes API ， 404
			if len(path) >= 5 && path[:5] == "/api/" {
				c.JSON(http.StatusNotFound, gin.H{"error": "API  "})
				return
			}
			// No index.html
			c.File("./static/index.html")
		})
	}

	// API（）
	api := r.Group("/api")
	api.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s Origin:%s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.Request.Header.Get("Origin"),
		)
	}))
	{
		// AuthenticationOff
		authRoutes := api.Group("/auth")
		{
			authRoutes.GET("/check-init", auth.CheckInit)
			authRoutes.POST("/init", auth.InitUser)
			authRoutes.POST("/login", auth.Login)
			authRoutes.POST("/refresh", auth.RefreshToken)
			authRoutes.POST("/logout", auth.Logout)
		}

		// Authentication
		protected := api.Group("") // ，
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/profile", auth.GetProfile)

			// Server Management
			serverRoutes := protected.Group("/servers")
			{
				serverRoutes.GET("", servers.GetServers)
				serverRoutes.POST("", servers.CreateServer)
				serverRoutes.GET("/:id", servers.GetServer)
				serverRoutes.PUT("/:id", servers.UpdateServer)
				serverRoutes.DELETE("/:id", servers.DeleteServer)
				serverRoutes.POST("/:id/start", servers.StartServer)
				serverRoutes.POST("/:id/stop", servers.StopServer)
				serverRoutes.POST("/:id/restart", servers.RestartServer)
				serverRoutes.POST("/:id/recreate", servers.RecreateContainer)
				serverRoutes.GET("/:id/rcon", servers.GetServerRCON)
				serverRoutes.GET("/:id/logs", servers.GetServerLogs)
			}

			// Image Management
			imageRoutes := protected.Group("/images")
			{
				imageRoutes.GET("/status", images.GetImageStatus)
				imageRoutes.POST("/pull", images.PullImage)
				imageRoutes.GET("/check-updates", images.CheckImageUpdates)
				imageRoutes.POST("/update", images.UpdateImage)
				imageRoutes.GET("/affected", images.GetAffectedServers)
			}

			// Plugins
			pluginRoutes := protected.Group("/plugins")
			{
				pluginRoutes.GET("", plugins.ListFiles)
				pluginRoutes.POST("/upload", plugins.UploadFile)
				pluginRoutes.DELETE("/delete", plugins.DeleteFile)
				pluginRoutes.POST("/rename", plugins.RenameFile)
				pluginRoutes.POST("/mkdir", plugins.CreateDir)
				pluginRoutes.GET("/download", plugins.DownloadFile)
				pluginRoutes.GET("/read", plugins.ReadFile)
				pluginRoutes.POST("/write", plugins.WriteFile)
				pluginRoutes.POST("/unzip", plugins.UnzipFile)
				pluginRoutes.GET("/zip-download", plugins.ZipDownload)
			}

			// Update status（Issue #3）
			updateRoutes := api.Group("/updates")
			{
				updateRoutes.GET("/:id/status", func(c *gin.Context) {
					serverID, err := strconv.ParseUint(c.Param("id"), 10, 32)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "None Server ID"})
						return
					}

					status, err := updateService.GetUpdateStatus(uint(serverID))
					if err != nil {
						c.JSON(http.StatusNotFound, gin.H{"error": "Update status "})
						return
					}

					c.JSON(http.StatusOK, gin.H{
						"message": "Operation successful",
						"data":    status,
					})
				})
			}

			// WebSocket 
			r.GET("/ws/updates/:id", hub.HandleWebSocket)
		}
	}
}
