package storage

import (
	"context"
	"io"
)

type Storage interface {
	// Upload uploads a file to storage and returns the file path
	Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)

	// Download retrieves a file from storage
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, path string) error

	// GetURL returns the public URL for a file
	GetURL(path string) string
}
