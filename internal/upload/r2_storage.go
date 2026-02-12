package upload

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type R2Storage struct {
	Client        *s3.Client
	PresignClient *s3.PresignClient
}

func (r *R2Storage) Upload(ctx context.Context, file FileInput, bucket string) (string, error) {
	sanitized := regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(file.Filename), "-")
	objectKey := fmt.Sprintf("%s-%s", uuid.New().String(), sanitized)

	contentType := file.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploadCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	_, err := r.Client.PutObject(uploadCtx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(objectKey),
		Body:        file.Reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("R2 PutObject failed: %w", err)
	}

	return objectKey, nil
}

func (r *R2Storage) Delete(ctx context.Context, bucket, objectKey string) error {
	deleteCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := r.Client.DeleteObject(deleteCtx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete R2 object: %w", err)
	}

	return nil
}

func (r *R2Storage) GenerateDownloadURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	presigned, err := r.PresignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presigned.URL, nil
}
