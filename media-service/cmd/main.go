package main

import (
	//"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Thedrogon/blogbish/media-service/internal/cache"
	"github.com/Thedrogon/blogbish/media-service/internal/handler"
	"github.com/Thedrogon/blogbish/media-service/internal/service"
	"github.com/Thedrogon/blogbish/media-service/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize MinIO storage
	storageProvider, err := storage.NewMinIOStorage()
	if err != nil {
		log.Fatalf("Failed to initialize MinIO storage: %v", err)
	}

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(
		fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		os.Getenv("REDIS_PASSWORD"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Redis cache: %v", err)
	}

	// Initialize services
	mediaService := service.NewMediaService(storageProvider, redisCache)

	// Initialize handlers
	mediaHandler := handler.NewMediaHandler(mediaService)

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(corsMiddleware())

	// Configure routes
	api := router.Group("/api/v1")
	{
		media := api.Group("/media")
		{
			media.POST("/upload", mediaHandler.Upload)
			media.GET("/:id", mediaHandler.GetMetadata)
			media.GET("/:id/download", mediaHandler.Download)
			media.PUT("/:id/metadata", mediaHandler.UpdateMetadata)
			media.DELETE("/:id", mediaHandler.Delete)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Media service started on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
