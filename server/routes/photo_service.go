package routes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mqtt-streaming-server/domain"
	"mqtt-streaming-server/utils"
)

// ErrPhotoNotFound is returned when a requested photo does not exist.
var ErrPhotoNotFound = errors.New("photo not found")

// PhotoService contains the business logic for photo operations.
type PhotoService struct {
	Repo domain.PhotoRepository
}

// NewPhotoService creates a new PhotoService backed by the given repository.
func NewPhotoService(repo domain.PhotoRepository) *PhotoService {
	return &PhotoService{Repo: repo}
}

// ListPhotos retrieves photos filtered by time range, optional full-text search,
// and optional device ID. Presigned download URLs are attached to each result.
func (s *PhotoService) ListPhotos(ctx context.Context, startUnix, endUnix int64, text, deviceID string) ([]*domain.Photo, error) {
	filters := map[string]any{
		"timestamp": map[string]any{
			"$gte": time.Unix(startUnix, 0),
			"$lte": time.Unix(endUnix, 0),
		},
	}
	if text != "" {
		filters["text"] = map[string]any{
			"$regex":   text,
			"$options": "i",
		}
	}
	if deviceID != "" {
		filters["device_id"] = deviceID
	}

	photos, err := s.Repo.GetPhotos(ctx, filters)
	if err != nil {
		return nil, err
	}

	for _, photo := range photos {
		keyName := fmt.Sprintf("photos/%d.%s", photo.Timestamp.Unix(), photo.ImageType)
		photo.PresignedURL = utils.GetPresignedURL(keyName)
	}

	return photos, nil
}

// DeletePhoto removes a photo from the database and object storage by ID.
// Returns ErrPhotoNotFound if the photo does not exist.
func (s *PhotoService) DeletePhoto(ctx context.Context, id string) error {
	photo, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return ErrPhotoNotFound
	}

	if err := s.Repo.Delete(ctx, id); err != nil {
		return err
	}

	keyName := fmt.Sprintf("photos/%d.%s", photo.Timestamp.Unix(), photo.ImageType)
	if err := utils.DeleteFromMinIO(keyName); err != nil {
		fmt.Printf("Warning: Could not delete object %s: %v\n", keyName, err)
	}

	return nil
}

// DeleteAllPhotos removes every photo from the database and clears the storage prefix.
// Returns the number of deleted database records.
func (s *PhotoService) DeleteAllPhotos(ctx context.Context) (int64, error) {
	count, err := s.Repo.DeleteAll(ctx)
	if err != nil {
		return 0, err
	}

	if err := utils.DeletePrefixFromMinIO("photos/"); err != nil {
		fmt.Printf("Warning: Could not delete all image objects: %v\n", err)
	}

	return count, nil
}
