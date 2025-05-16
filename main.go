package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	router *chi.Mux
	logger *log.Logger
}

func NewServer() *Server {
	r := chi.NewRouter()
	logger := log.New(os.Stdout, "[BLOGBISH] ", log.LstdFlags)

	// Basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	return &Server{
		router: r,
		logger: logger,
	}
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Auth Service Routes
	s.router.Route("/auth", func(r chi.Router) {
		r.Post("/register", s.proxyRequest("http://auth-service:8080/register"))
		r.Post("/login", s.proxyRequest("http://auth-service:8080/login"))
		r.Post("/refresh", s.proxyRequest("http://auth-service:8080/refresh"))
		r.Post("/logout", s.proxyRequest("http://auth-service:8080/logout"))
	})

	// Post Service Routes
	s.router.Route("/posts", func(r chi.Router) {
		r.Get("/", s.proxyRequest("http://post-service:8081/posts"))
		r.Post("/", s.proxyRequest("http://post-service:8081/posts"))
		r.Get("/{id}", s.proxyRequest("http://post-service:8081/posts/{id}"))
		r.Put("/{id}", s.proxyRequest("http://post-service:8081/posts/{id}"))
		r.Delete("/{id}", s.proxyRequest("http://post-service:8081/posts/{id}"))
	})

	// Comment Service Routes
	s.router.Route("/comments", func(r chi.Router) {
		r.Get("/", s.proxyRequest("http://comment-service:8083/comments"))
		r.Post("/", s.proxyRequest("http://comment-service:8083/comments"))
		r.Get("/{id}", s.proxyRequest("http://comment-service:8083/comments/{id}"))
		r.Put("/{id}", s.proxyRequest("http://comment-service:8083/comments/{id}"))
		r.Delete("/{id}", s.proxyRequest("http://comment-service:8083/comments/{id}"))
		r.Post("/{id}/like", s.proxyRequest("http://comment-service:8083/comments/{id}/like"))
		r.Post("/{id}/report", s.proxyRequest("http://comment-service:8083/comments/{id}/report"))
		r.Put("/{id}/moderate", s.proxyRequest("http://comment-service:8083/comments/{id}/moderate"))
		r.Get("/ws", s.proxyRequest("http://comment-service:8083/ws"))
	})

	// Media Service Routes
	s.router.Route("/media", func(r chi.Router) {
		r.Post("/upload", s.proxyRequest("http://media-service:8082/upload"))
		r.Get("/{id}", s.proxyRequest("http://media-service:8082/media/{id}"))
		r.Delete("/{id}", s.proxyRequest("http://media-service:8082/media/{id}"))
	})

	// Search Service Routes
	s.router.Route("/search", func(r chi.Router) {
		r.Get("/posts", s.proxyRequest("http://search-service:8084/search/posts"))
		r.Get("/users", s.proxyRequest("http://search-service:8084/search/users"))
	})
}

func (s *Server) proxyRequest(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a new request
		proxyReq, err := http.NewRequest(r.Method, target, r.Body)
		if err != nil {
			s.logger.Printf("Error creating proxy request: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		// Copy URL parameters
		proxyReq.URL.RawQuery = r.URL.RawQuery

		// Send the request
		client := &http.Client{
			Timeout: time.Second * 30,
		}
		resp, err := client.Do(proxyReq)
		if err != nil {
			s.logger.Printf("Error sending proxy request: %v", err)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// Copy status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		if _, err := http.MaxBytesReader(w, resp.Body, 32<<20).WriteTo(w); err != nil {
			s.logger.Printf("Error copying response body: %v", err)
		}
	}
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8000, "API Gateway port")
	flag.Parse()

	server := NewServer()
	server.setupRoutes()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      server.router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		server.logger.Printf("API Gateway starting on port %d", port)
		serverErrors <- httpServer.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		server.logger.Fatalf("Error starting server: %v", err)

	case sig := <-shutdown:
		server.logger.Printf("Start shutdown: %v", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := httpServer.Shutdown(ctx); err != nil {
			server.logger.Printf("Graceful shutdown did not complete in %v: %v", 30*time.Second, err)
			if err := httpServer.Close(); err != nil {
				server.logger.Fatalf("Could not stop http server: %v", err)
			}
		}
	}
}
