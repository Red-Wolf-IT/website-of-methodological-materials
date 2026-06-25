package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"website-of-methodological-materials/internal/handlers"
	appmiddleware "website-of-methodological-materials/internal/middleware"
)

type Config struct {
	AdminToken string
}

func New(
	cfg Config,
	healthHandler *handlers.HealthHandler,
	manualHandler *handlers.ManualHandler,
	tagHandler *handlers.TagHandler,
	fileHandler *handlers.FileHandler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.CORS)
	r.Use(appmiddleware.Logger)
	r.Use(appmiddleware.Recover)

	r.Get("/health", healthHandler.ServeHTTP)

	r.Get("/tags", tagHandler.List)
	r.Post("/tags", tagHandler.Create)

	r.Post("/manuals", manualHandler.Create)
	r.Get("/manuals", manualHandler.List)
	r.Get("/manuals/{id}", manualHandler.GetByID)
	r.Post("/manuals/{id}/tags", manualHandler.AttachTags)

	r.Get("/uploads/{filename}", fileHandler.Serve)

	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.AdminAuth(cfg.AdminToken))

		r.Put("/manuals/{id}", manualHandler.Update)
		r.Delete("/manuals/{id}", manualHandler.Delete)
		r.Post("/manuals/{id}/attachment", manualHandler.UploadAttachment)
	})

	return r
}
