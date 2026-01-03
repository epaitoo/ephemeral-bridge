package data

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ConfigEnv struct {
	Port        int
	AppEnv      string
	DatabaseURL string
}

func LoadConfig() (*ConfigEnv, error) {
	err := godotenv.Load(".env")

	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("PORT is not set in .env file")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("could not convert PORT to int")
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set in .env file")
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		return nil, fmt.Errorf("APP_ENV is not set in .env file")
	}

	return &ConfigEnv{
		Port:        portNum,
		AppEnv:      appEnv,
		DatabaseURL: dbUrl,
	}, nil

}
