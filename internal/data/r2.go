package data

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewR2Client(cfg *ConfigEnv) (*s3.Client, error) {
	r2Cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyId, cfg.R2SecretAccessKey, ""),
		),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	client := s3.NewFromConfig(r2Cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.R2s3Api)
	})

	return client, nil
}
