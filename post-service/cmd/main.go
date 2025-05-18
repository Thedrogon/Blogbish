package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Thedrogon/blogbish/post-service/internal/cache"
	"github.com/Thedrogon/blogbish/post-service/internal/handler"
	"github.com/Thedrogon/blogbish/post-service/internal/repository"
	"github.com/Thedrogon/blogbish/post-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	// Initialize handlers
	postHandler := handler.NewPostHandler(postService)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Route("/posts", func(r chi.Router) {
		r.Get("/", postHandler.List)
		r.Post("/", postHandler.Create)
		r.Get("/{id}", postHandler.Get)
		r.Put("/{id}", postHandler.Update)
		r.Delete("/{id}", postHandler.Delete)
	})

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", categoryHandler.List)
		r.Post("/", categoryHandler.Create)
		r.Get("/{slug}", categoryHandler.Get)
		r.Put("/{slug}", categoryHandler.Update)
		r.Delete("/{slug}", categoryHandler.Delete)
	})

	// Start server
	port := getEnv("PORT", "8081")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Channel for server errors
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Printf("Post service starting on port %s", port)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown or error
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Start shutdown: %v", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in %v: %v", 30*time.Second, err)
			if err := srv.Close(); err != nil {
				log.Fatalf("Could not stop http server: %v", err)
			}
		}
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
