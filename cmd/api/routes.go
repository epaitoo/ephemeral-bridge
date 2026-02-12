package main

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func (app *application) routes() chi.Router {
	//setup chi router
	r := app.chiRouter
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// r.NotFound(app.notFoundResponse)
	r.Get("/v1/healthcheck", app.healthcheckHandler)

	r.Group(func(r chi.Router) {
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
