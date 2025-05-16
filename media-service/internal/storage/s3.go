package storage

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	region     string
	baseURL    string
}

func NewS3Storage(client *s3.Client, bucketName, region string) *S3Storage {
	baseURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucketName, region)
	return &S3Storage{
		client:     client,
		bucketName: bucketName,
		region:     region,
		baseURL:    baseURL,
	}
}

func (s *S3Storage) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return filename, nil
}

func (s *S3Storage) Download(ctx context.Context, filepath string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(filepath),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	return result.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, filepath string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(filepath),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

func (s *S3Storage) GetURL(filepath string) string {
	return path.Join(s.baseURL, filepath)
}
