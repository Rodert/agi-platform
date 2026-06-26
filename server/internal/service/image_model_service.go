package service

import (
	"context"

	"agi-platform/server/internal/model"
	"agi-platform/server/internal/repository"
)

type ImageModelService interface {
	ListEnabled(ctx context.Context) ([]model.ImageModel, error)
}

type imageModelService struct {
	models repository.ImageModelRepository
}

func NewImageModelService(models repository.ImageModelRepository) ImageModelService {
	return &imageModelService{models: models}
}

func (s *imageModelService) ListEnabled(ctx context.Context) ([]model.ImageModel, error) {
	return s.models.ListEnabled(ctx)
}
