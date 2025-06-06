package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Thedrogon/blogbish/Internals/routes" 
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

		// Copy response body using io.Copy with MaxBytesReader
		limitedReader := http.MaxBytesReader(w, resp.Body, 32<<20)
		if _, err := io.Copy(w, limitedReader); err != nil {
			s.logger.Printf("Error copying response body: %v", err)
		}
	}
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8000, "API Gateway port")
	flag.Parse()

	server := NewServer()



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
