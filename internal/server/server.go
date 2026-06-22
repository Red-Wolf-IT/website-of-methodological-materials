package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"website-of-methodological-materials/internal/handlers"
	appmiddleware "website-of-methodological-materials/internal/middleware"
)

func New(healthHandler *handlers.HealthHandler, manualHandler *handlers.ManualHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.Logger)
	r.Use(appmiddleware.Recover)

	r.Get("/health", healthHandler.ServeHTTP)

	r.Post("/manuals", manualHandler.Create)
	r.Get("/manuals/{id}", manualHandler.GetByID)

	return r
}
