package main

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	mw "github.com/epaitoo/ephermalbridge/internal/middleware"
)

func (app *application) routes() chi.Router {
	r := app.chiRouter

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(mw.RateLimitMiddleware(10, 20))

	if len(app.configEnv.CORSAllowedOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   app.configEnv.CORSAllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	// Public routes
	r.Get("/v1/healthcheck", app.healthcheckHandler)

	// Auth routes: Cloudflare verification only (no API key/session needed)
	r.Group(func(r chi.Router) {
		r.Use(mw.CloudflareMiddleware(app.authConfig, app.cfVerifier))
		r.Get("/v1/auth/session", app.createSessionHandler)
		r.Post("/v1/auth/session", app.createSessionHandler)
		r.Post("/v1/auth/logout", app.logoutHandler)
	})

	// Protected routes: both Cloudflare + API key/session
	r.Group(func(r chi.Router) {
		r.Use(mw.AuthMiddleware(app.authConfig, app.cfVerifier))

		r.Get("/v1/texts/{id}", app.showTextsHandler)
		r.Get("/v1/texts", app.getAllTextsHandler)
		r.Post("/v1/texts", app.createTextHandler)
		r.Patch("/v1/texts/{id}", app.updateTextHandler)
		r.Delete("/v1/texts/{id}", app.deleteTextHandler)

		r.Post("/v1/files", app.uploadFilesHandler)
		r.Get("/v1/files", app.getAllFilesHandler)
		r.Get("/v1/files/{id}", app.getFileHandler)
		r.Get("/v1/files/{id}/download", app.downloadFileHandler)
		r.Delete("/v1/files/{id}", app.deleteFileHandler)
		r.Post("/v1/files/cleanup", app.deleteExpiredFilesHandler)
	})

	return r
}
