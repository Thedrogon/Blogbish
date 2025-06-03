package routes

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (s *Server) SetupRoutes() {

	
	// Health check
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Auth Service Routes
	s.router.Route("/auth", func(r chi.Router) {
		r.Post("/register", s.proxyRequest("http://auth  -service:8080/register"))
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