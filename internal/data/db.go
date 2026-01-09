package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPool(connStr string) (*pgxpool.Pool, error) {
	const (
		defaultMaxConns        = 10
		defaultMinConns        = 0
		defaultMaxConnLifetime = time.Hour
		defaultMaxConnIdleTime = 30 * time.Minute
		defaultHealthCheck     = time.Minute
		defaultConnectTimeout  = 5 * time.Second
	)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection config: %w", err)
	}

	config.MaxConns = defaultMaxConns
	config.MinConns = defaultMinConns
	config.MaxConnLifetime = defaultMaxConnLifetime
	config.MaxConnIdleTime = defaultMaxConnIdleTime
	config.HealthCheckPeriod = defaultHealthCheck
	config.ConnConfig.ConnectTimeout = defaultConnectTimeout

	// Create pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	if err = pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	slog.Info("Successfully connected to database")
	return pool, nil
}

// NewDB creates a standard sql.DB connection for migrations
func NewDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	// Verify connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return db, nil
}
