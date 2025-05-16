package service

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/Thedrogon/blogbish/media-service/internal/cache"
	"github.com/Thedrogon/blogbish/media-service/internal/models"
	"github.com/Thedrogon/blogbish/media-service/internal/storage"
)

var (
	ErrNotFound  = errors.New("media not found")
	ErrForbidden = errors.New("forbidden")
)

type MediaService struct {
	storage storage.Storage
	cache   cache.Cache
}

func NewMediaService(storage storage.Storage, cache cache.Cache) *MediaService {
	return &MediaService{
		storage: storage,
		cache:   cache,
	}
}

func (s *MediaService) UploadFile(ctx context.Context, file io.Reader, filename string, contentType string, upload *models.MediaUpload) (*models.MediaResponse, error) {
	// Upload file to storage
	path, err := s.storage.Upload(ctx, file, filename, contentType)
	if err != nil {
		return nil, err
	}

	// Create media record
	media := &models.Media{
		ID:          filename, // Using filename as ID since it's UUID
		UserID:      upload.UserID,
		Filename:    filename,
		ContentType: contentType,
		Path:        path,
		URL:         s.storage.GetURL(path),
		Metadata: models.Metadata{
			Title:       upload.Title,
			Description: upload.Description,
			AltText:     upload.AltText,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Cache the media metadata
	if err := s.cache.SetMedia(ctx, media); err != nil {
		// Log error but don't fail the upload
		// TODO: Add proper logging
	}

	return media.ToResponse(), nil
}

func (s *MediaService) GetFile(ctx context.Context, id string) (*models.Media, error) {
	// Try to get from cache first
	media, err := s.cache.GetMedia(ctx, id)
	if err == nil {
		return media, nil
	}

	// If not in cache, return not found
	// In a real implementation, we would have a database to check
	return nil, ErrNotFound
}

func (s *MediaService) DownloadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.storage.Download(ctx, path)
}

func (s *MediaService) DeleteFile(ctx context.Context, id string, userID int64) error {
	// Get file metadata
	media, err := s.GetFile(ctx, id)
	if err != nil {
		return err
	}

	// Check ownership
	if media.UserID != userID {
		return ErrForbidden
	}

	// Delete from storage
	if err := s.storage.Delete(ctx, media.Path); err != nil {
		return err
	}

	// Delete from cache
	if err := s.cache.DeleteMedia(ctx, id); err != nil {
		// Log error but don't fail the deletion
		// TODO: Add proper logging
	}

	return nil
}

func (s *MediaService) UpdateMetadata(ctx context.Context, id string, userID int64, metadata *models.Metadata) (*models.MediaResponse, error) {
	// Get file metadata
	media, err := s.GetFile(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if media.UserID != userID {
		return nil, ErrForbidden
	}

	// Update metadata
	media.Metadata = *metadata
	media.UpdatedAt = time.Now()

	// Update cache
	if err := s.cache.SetMedia(ctx, media); err != nil {
		// Log error but don't fail the update
		// TODO: Add proper logging
	}

	return media.ToResponse(), nil
}
