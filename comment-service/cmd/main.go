package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Thedrogon/blogbish/comment-service/internal/handler"
	"github.com/Thedrogon/blogbish/comment-service/internal/repository"
	"github.com/Thedrogon/blogbish/comment-service/internal/service"
	"github.com/Thedrogon/blogbish/comment-service/internal/websocket"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize PostgreSQL connection
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize repository
	commentRepo := repository.NewCommentRepository(db)

	// Initialize service
	commentService := service.NewCommentService(commentRepo, hub)

	// Initialize handler
	commentHandler := handler.NewCommentHandler(commentService, hub)

	// Initialize router
	router := gin.Default()

	// Register routes
	router.POST("/comments", commentHandler.CreateComment)
	router.GET("/comments/:id", commentHandler.GetComment)
	router.PUT("/comments/:id", commentHandler.UpdateComment)
	router.DELETE("/comments/:id", commentHandler.DeleteComment)
	router.GET("/comments", commentHandler.ListComments)
	router.POST("/comments/:id/like", commentHandler.LikeComment)
	router.POST("/comments/:id/report", commentHandler.ReportComment)
	router.PUT("/comments/:id/moderate", commentHandler.ModerateComment)
	router.GET("/ws", commentHandler.WebSocket)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
