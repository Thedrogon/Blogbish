package app

import (
	"log"

	"github.com/go-chi/chi"
)

type Application struct {
	router *chi.Mux
	logger *log.Logger
}

func NewApplication() (*Application , error) {
	
}