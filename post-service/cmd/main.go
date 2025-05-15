package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Thedrogon/blogbish/post-service/internal/cache"
	"github.com/Thedrogon/blogbish/post-service/internal/repository"
	"github.com/Thedrogon/blogbish/post-service/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	// Database configuration
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "blogbish")

	// Redis configuration
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")

	// Connect to PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(
		fmt.Sprintf("%s:%s", redisHost, redisPort),
		redisPassword,
	)
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}

	// Initialize repositories
	postRepo := repository.NewPostgresPostRepository(db)
	categoryRepo := repository.NewPostgresCategoryRepository(db)

	// Initialize services with cache
	postService := service.NewPostService(postRepo, categoryRepo, redisCache)
	categoryService := service.NewCategoryService(categoryRepo, redisCache)

	// TODO: Initialize and start HTTP server
	// This part would typically involve setting up your HTTP router,
	// handlers, middleware, etc.

	log.Println("Post service started successfully")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
