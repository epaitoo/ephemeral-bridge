package data

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ConfigEnv struct {
	Port              int
	AppEnv            string
	DatabaseURL       string
	R2AccessKeyId     string
	R2SecretAccessKey string
	R2AccountId       string
	R2BucketName      string
	R2TokenValue      string
	R2s3Api           string
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

	// R2 Configs
	r2AccessKeyId := os.Getenv("R2_ACCESS_KEY_ID")
	if r2AccessKeyId == "" {
		return nil, fmt.Errorf("R2_ACCESS_KEY_ID is not set in .env file")
	}

	r2SecretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	if r2SecretAccessKey == "" {
		return nil, fmt.Errorf("R2_SECRET_ACCESS_KEY is not set in .env file")
	}

	r2AccountId := os.Getenv("R2_ACCOUNT_ID")
	if r2AccountId == "" {
		return nil, fmt.Errorf("R2_ACCOUNT_ID is not set in .env file")
	}

	r2BucketName := os.Getenv("R2_BUCKET_NAME")
	if r2BucketName == "" {
		return nil, fmt.Errorf("R2_BUCKET_NAME is not set in .env file")
	}

	r2TokenValue := os.Getenv("R2_TOKEN_VALUE")
	if r2TokenValue == "" {
		return nil, fmt.Errorf("R2_TOKEN_VALUE is not set in .env file")
	}

	r2S3Api := os.Getenv("R2_S3_API")
	if r2S3Api == "" {
		return nil, fmt.Errorf("R2_S3_API is not set in .env file")
	}

	return &ConfigEnv{
		Port:              portNum,
		AppEnv:            appEnv,
		DatabaseURL:       dbUrl,
		R2AccessKeyId:     r2AccessKeyId,
		R2SecretAccessKey: r2SecretAccessKey,
		R2AccountId:       r2AccountId,
		R2BucketName:      r2BucketName,
		R2TokenValue:      r2TokenValue,
		R2s3Api:           r2S3Api,
	}, nil

}
