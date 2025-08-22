package app

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Application struct {
	router *chi.Mux
	logger *log.Logger
}

func NewApplication() (*Application, error) {
	router := chi.NewRouter()
	logger := log.New(os.Stdout, "[BLOGBISH] ", log.LstdFlags)

	// Basic middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))

	app := &Application{
		router: router,
		logger: logger,
	}

	// Setup routes
	app.SetupRoutes()

	return app, nil
}

func (s *Application) SetupRoutes() {
	// Health check
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			s.logger.Printf("Error writing response: %v", err)
		}
	})
	// Auth Service Routes
	s.router.Route("/auth", func(r chi.Router) {
		r.Post("/register", s.proxyRequest("http://auth-service:8080/auth/register"))
		r.Post("/login", s.proxyRequest("http://auth-service:8080/auth/login"))
		r.Get("/me", s.proxyRequest("http://auth-service:8080/auth/me"))
	})

	// Post Service Routes
	s.router.Route("/posts", func(r chi.Router) {
		r.Get("/", s.proxyRequest("http://post-service:8081/posts"))
		r.Post("/", s.proxyRequest("http://post-service:8081/posts"))
		r.Get("/{id}", s.proxyRequest("http://post-service:8081/posts/{id}"))
		r.Put("/{id}", s.proxyRequest("http://post-service:8081/posts/{id}"))
		r.Delete("/{id}", s.proxyRequest("http://post-service:8081/posts/{id}"))
	})

	// Category Service Routes
	s.router.Route("/categories", func(r chi.Router) {
		r.Get("/", s.proxyRequest("http://post-service:8081/categories"))
		r.Post("/", s.proxyRequest("http://post-service:8081/categories"))
		r.Get("/{slug}", s.proxyRequest("http://post-service:8081/categories/{slug}"))
		r.Put("/{slug}", s.proxyRequest("http://post-service:8081/categories/{slug}"))
		r.Delete("/{slug}", s.proxyRequest("http://post-service:8081/categories/{slug}"))
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
		r.Post("/upload", s.proxyRequest("http://media-service:8082/api/v1/media/upload"))
		r.Get("/{id}", s.proxyRequest("http://media-service:8082/api/v1/media/{id}"))
		r.Get("/{id}/download", s.proxyRequest("http://media-service:8082/api/v1/media/{id}/download"))
		r.Put("/{id}/metadata", s.proxyRequest("http://media-service:8082/api/v1/media/{id}/metadata"))
		r.Delete("/{id}", s.proxyRequest("http://media-service:8082/api/v1/media/{id}"))
	})

	// Search Service Routes
	s.router.Route("/search", func(r chi.Router) {
		r.Post("/", s.proxyRequest("http://search-service:8084/search"))
		r.Post("/suggest", s.proxyRequest("http://search-service:8084/suggest"))
		r.Post("/index/post", s.proxyRequest("http://search-service:8084/index/post"))
		r.Post("/index/comment", s.proxyRequest("http://search-service:8084/index/comment"))
	})
}

func (s *Application) proxyRequest(targetURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target, err := url.Parse(targetURL)
		if err != nil {
			s.logger.Printf("Error parsing target URL: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
	}
}

func (s *Application) GetRouter() *chi.Mux {
	return s.router
}
