package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/epaitoo/ephermalbridge/internal/auth"
	"github.com/epaitoo/ephermalbridge/internal/config"
	"github.com/epaitoo/ephermalbridge/internal/data"
	"github.com/epaitoo/ephermalbridge/internal/upload"
	"github.com/go-chi/chi/v5"
)

const version string = "1.0.0"

type application struct {
	chiRouter   *chi.Mux
	configEnv   data.ConfigEnv
	authConfig  *config.AuthConfig
	cfVerifier  *auth.CloudflareVerifier
	logger      *slog.Logger
	models      data.Models
	r2Client    *s3.Client
	storage     *upload.R2Storage
	coordinator *upload.UploadCoordinator
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

	// Load auth config
	authCfg, err := config.LoadAuthConfig()
	if err != nil {
		logger.Error("auth config error", "error", err.Error())
		os.Exit(1)
	}

	var cfVerifier *auth.CloudflareVerifier
	if !authCfg.SkipCloudflareAuth {
		cfVerifier = auth.NewCloudflareVerifier(authCfg.CloudflareTeamDomain, authCfg.CloudflareAudience)
		logger.Info("Cloudflare Access verification enabled")
	} else {
		logger.Info("Cloudflare Access verification skipped (development mode)")
	}

	// Run migrations
	logger.Info("Running database migrations...")
	db, err := data.NewDB(configEnv.DatabaseURL)
	if err != nil {
		logger.Error("DB connection for migrations failed", "error", err.Error())
		os.Exit(1)
	}

	if err := data.RunMigrations(db); err != nil {
		logger.Error("Failed to run migrations", "error", err.Error())
		db.Close()
		os.Exit(1)
	}
	db.Close()
	logger.Info("Database migrations completed successfully")

	// load DB from Config
	dbpool, err := data.NewPool(configEnv.DatabaseURL)
	if err != nil {
		logger.Error("DB Connections Error ", "error", err.Error())
		os.Exit(1)
	}

	defer dbpool.Close()

	// Initialize R2 client
	r2Client, err := data.NewR2Client(configEnv)
	if err != nil {
		logger.Error("R2 client initialization failed", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("R2 client initialized successfully")

	storage := &upload.R2Storage{
		Client:        r2Client,
		PresignClient: s3.NewPresignClient(r2Client),
	}

	models := data.NewModels(dbpool)

	coordinator := &upload.UploadCoordinator{
		Storage:    storage,
		Repository: &upload.PGFileRepository{Model: models.Files},
		Logger:     logger,
		BucketName: configEnv.R2BucketName,
	}

	//setup chi router
	r := chi.NewRouter()

	//app struct
	app := application{
		configEnv:   *configEnv,
		authConfig:  authCfg,
		cfVerifier:  cfVerifier,
		logger:      logger,
		chiRouter:   r,
		models:      models,
		r2Client:    r2Client,
		storage:     storage,
		coordinator: coordinator,
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
