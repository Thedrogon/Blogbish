package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client     *minio.Client
	bucketName string
}

func NewMinIOStorage() (*MinIOStorage, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	// Create bucket if it doesn't exist
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	return &MinIOStorage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *MinIOStorage) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// Generate a unique filename
	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	contentType := file.Header.Get("Content-Type")

	// Upload the file to MinIO
	_, err = s.client.PutObject(ctx, s.bucketName, filename, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	return filename, nil
}

func (s *MinIOStorage) Download(ctx context.Context, filename string) ([]byte, string, error) {
	// Get the object
	object, err := s.client.GetObject(ctx, s.bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object: %v", err)
	}
	defer object.Close()

	// Get object info
	info, err := object.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object info: %v", err)
	}

	// Read the object data
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, object); err != nil {
		return nil, "", fmt.Errorf("failed to read object data: %v", err)
	}

	return buffer.Bytes(), info.ContentType, nil
}

func (s *MinIOStorage) Delete(ctx context.Context, filename string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}
	return nil
}

func (s *MinIOStorage) GetURL(filename string) string {
	return fmt.Sprintf("/api/v1/media/%s", filename)
} 