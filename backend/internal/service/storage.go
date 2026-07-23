package service

import (
	"context"

	"github.com/javapub/agi-platform-backend/internal/objectstorage"
)

// StorageService is the application-facing facade over the active object store.
// Callers select a resource type; provider configuration stays outside business logic.
type StorageService struct{ manager *objectstorage.Manager }

func NewStorageService(manager *objectstorage.Manager) *StorageService {
	return &StorageService{manager: manager}
}

func (s *StorageService) UploadBase64Image(ctx context.Context, encoded string) (*objectstorage.StoredObject, error) {
	return s.manager.UploadBase64(ctx, "reference", encoded)
}

func (s *StorageService) UploadGeneratedBase64(ctx context.Context, resourceType, encoded string) (*objectstorage.StoredObject, error) {
	return s.manager.UploadBase64(ctx, resourceType, encoded)
}

func (s *StorageService) UploadGeneratedURL(ctx context.Context, resourceType, sourceURL string) (*objectstorage.StoredObject, error) {
	return s.manager.UploadFromURL(ctx, resourceType, sourceURL)
}
