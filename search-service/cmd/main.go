package main

import (
	"log"
	"os"

	"github.com/Thedrogon/blogbish/search-service/internal/handler"
	"github.com/Thedrogon/blogbish/search-service/internal/repository"
	"github.com/Thedrogon/blogbish/search-service/internal/service"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Initialize Elasticsearch client
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTICSEARCH_URL")},
	})
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})

	// Initialize repository
	searchRepo := repository.NewElasticsearchRepository(esClient)

	// Initialize service
	searchService := service.NewSearchService(searchRepo, redisClient)

	// Initialize handler
	searchHandler := handler.NewSearchHandler(searchService)

	// Initialize router
	router := gin.Default()

	// Register routes
	router.POST("/search", searchHandler.Search)
	router.POST("/suggest", searchHandler.Suggest)
	router.POST("/index/post", searchHandler.IndexPost)
	router.POST("/index/comment", searchHandler.IndexComment)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
