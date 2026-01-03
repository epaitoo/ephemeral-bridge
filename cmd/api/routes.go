package main

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func (app *application) routes() chi.Router {
	//setup chi router
	r := app.chiRouter
	r.Use(middleware.Logger) // <--<< Logger should come before Recoverer
	r.Use(middleware.Recoverer)

	// r.NotFound(app.notFoundResponse)
	r.Get("/v1/healthcheck", app.healthcheckHandler)

	return r
}
