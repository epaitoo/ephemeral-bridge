package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/epaitoo/ephermalbridge/internal/data"
	"github.com/go-chi/chi/v5"
)

const version string = "1.0.0"

type application struct {
	chiRouter *chi.Mux
	configEnv data.ConfigEnv
	logger    *slog.Logger
}

func main() {

	// handle shutdown
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	configEnv, err := data.LoadConfig()

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1) // exit if config fails
	}

	flag.IntVar(&configEnv.Port, "port", configEnv.Port, "API Server Port")
	flag.StringVar(&configEnv.AppEnv, "environment", configEnv.AppEnv, "Environment (development|staging|production)")
	flag.Parse()

	//setup chi router
	r := chi.NewRouter()

	//app struct
	app := application{
		configEnv: *configEnv,
		logger:    logger,
		chiRouter: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server
	if err := app.serve(ctx); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info("application stopped gracefully")

}
