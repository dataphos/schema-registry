// Package server contains the Schema registry REST Server configuration and start-up functions.
package server

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"time"

	_ "github.com/dataphos/aquarium-janitor-standalone-sr/docs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// New sets up the schema registry endpoints.
func New(h *Handler) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.StripSlashes)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	router.Use(RequestLogger(h.log))

	router.Route("/schemas", func(router chi.Router) {
		router.Get("/", h.GetSchemas)
		router.Post("/", h.PostSchema)
		router.Get("/all", h.GetAllSchemas)

		router.Route("/{id}", func(router chi.Router) {
			router.Delete("/", h.DeleteSchema)
			router.Put("/", h.PutSchema)

			router.Route("/versions", func(router chi.Router) {
				router.Get("/", h.GetSchemaVersionsById)
				router.Get("/latest", h.GetLatestSchemaVersionById)
				router.Get("/all", h.GetAllSchemaVersionsById)

				router.Route("/{version}", func(router chi.Router) {
					router.Get("/", h.GetSchemaVersionByIdAndVersion)
					router.Delete("/", h.DeleteSchemaVersion)

					router.Route("/spec", func(router chi.Router) {
						router.Get("/", h.GetSpecificationByIdAndVersion)
					})
				})
			})
		})

		router.Get("/search", h.SearchSchemas)
	})

	router.Get("/health", h.HealthCheck)

	router.Post("/check/compatibility", h.SchemaCompatibility)
	router.Get("/check/compatibility/health", h.HealthCheck)

	router.Post("/check/validity", h.SchemaValidity)
	router.Get("/check/validity/health", h.HealthCheck)

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), //The url pointing to API definition
	))

	return router
}
