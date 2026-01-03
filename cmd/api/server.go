// server.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (app *application) serve(ctx context.Context) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.configEnv.Port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Run server in a goroutine
	go func() {
		app.logger.Info("starting server...", "addr", srv.Addr, "environment", app.configEnv.AppEnv)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Error("server error", "error", err)
		}
	}()

	// Block until context is canceled
	<-ctx.Done()

	// Shutdown gracefully when ctx is canceled
	app.logger.Info("shutting down server...", "addr", srv.Addr)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("could not gracefully shutdown the server: %w", err)
	}

	app.logger.Info("server stopped", "addr", srv.Addr)
	return nil
}
