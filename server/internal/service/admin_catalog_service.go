package service

import (
	"context"
	"encoding/json"

	"agi-platform/server/internal/model"
	"agi-platform/server/internal/repository"

	"gorm.io/datatypes"
)

type AdminCatalogService interface {
	ListProviders(ctx context.Context, limit int, offset int) ([]model.Provider, error)
	CreateProvider(ctx context.Context, req ProviderRequest) (*model.Provider, error)
	UpdateProvider(ctx context.Context, id uint64, req ProviderRequest) error
	ListProviderKeys(ctx context.Context, providerID uint64) ([]model.ProviderKey, error)
	CreateProviderKey(ctx context.Context, providerID uint64, req ProviderKeyRequest) (*model.ProviderKey, error)
	DeleteProviderKey(ctx context.Context, id uint64) error
	ListImageModels(ctx context.Context, limit int, offset int) ([]model.ImageModel, error)
	CreateImageModel(ctx context.Context, req ImageModelRequest) (*model.ImageModel, error)
	UpdateImageModel(ctx context.Context, id uint64, req ImageModelRequest) error
	ListImageModelRoutes(ctx context.Context, modelID uint64) ([]model.ImageModelRoute, error)
	CreateImageModelRoute(ctx context.Context, modelID uint64, req ImageModelRouteRequest) (*model.ImageModelRoute, error)
	UpdateImageModelRoute(ctx context.Context, id uint64, req ImageModelRouteRequest) error
}

type ProviderRequest struct {
	Code           string
	Name           string
	Type           string
	BaseURL        string
	Enabled        bool
	TimeoutSeconds int
	RetryCount     int
	Priority       int
	DailyLimit     *int64
	Remark         string
}

type ProviderKeyRequest struct {
	Name       string
	APIKey     string
	Status     string
	Weight     int
	DailyLimit *int64
}

type ImageModelRequest struct {
	Code                string
	DisplayName         string
	Description         string
	CoverURL            string
	PriceCredits        int64
	SupportedSizes      []string
	SupportTextToImage  bool
	SupportImageToImage bool
	SupportEdit         bool
	MaxImagesPerRequest int
	AutoRefundOnFailure bool
	Enabled             bool
	Recommended         bool
	SortOrder           int
}

type ImageModelRouteRequest struct {
	ProviderID        uint64
	ProviderModelName string
	Enabled           bool
	Priority          int
	Weight            int
	ExtraConfig       map[string]interface{}
}

type adminCatalogService struct {
	repos repository.Repositories
}

func NewAdminCatalogService(repos repository.Repositories) AdminCatalogService {
	return &adminCatalogService{repos: repos}
}

func (s *adminCatalogService) ListProviders(ctx context.Context, limit int, offset int) ([]model.Provider, error) {
	limit, offset = normalizePage(limit, offset)
	return s.repos.Providers.List(ctx, limit, offset)
}

func (s *adminCatalogService) CreateProvider(ctx context.Context, req ProviderRequest) (*model.Provider, error) {
	if req.Code == "" || req.Name == "" || req.Type == "" {
		return nil, ErrInvalidRequest
	}
	provider := &model.Provider{
		Code:           req.Code,
		Name:           req.Name,
		Type:           req.Type,
		BaseURL:        req.BaseURL,
		Enabled:        req.Enabled,
		TimeoutSeconds: defaultInt(req.TimeoutSeconds, 60),
		RetryCount:     req.RetryCount,
		Priority:       defaultInt(req.Priority, 100),
		DailyLimit:     req.DailyLimit,
		Remark:         req.Remark,
	}
	if err := s.repos.Providers.Create(ctx, provider); err != nil {
		return nil, err
	}
	return provider, nil
}

func (s *adminCatalogService) UpdateProvider(ctx context.Context, id uint64, req ProviderRequest) error {
	if id == 0 || req.Code == "" || req.Name == "" || req.Type == "" {
		return ErrInvalidRequest
	}
	return s.repos.Providers.Update(ctx, id, map[string]interface{}{
		"code":            req.Code,
		"name":            req.Name,
		"type":            req.Type,
		"base_url":        req.BaseURL,
		"enabled":         req.Enabled,
		"timeout_seconds": defaultInt(req.TimeoutSeconds, 60),
		"retry_count":     req.RetryCount,
		"priority":        defaultInt(req.Priority, 100),
		"daily_limit":     req.DailyLimit,
		"remark":          req.Remark,
	})
}

func (s *adminCatalogService) ListProviderKeys(ctx context.Context, providerID uint64) ([]model.ProviderKey, error) {
	return s.repos.Providers.ListKeys(ctx, providerID)
}

func (s *adminCatalogService) CreateProviderKey(ctx context.Context, providerID uint64, req ProviderKeyRequest) (*model.ProviderKey, error) {
	if providerID == 0 || req.APIKey == "" {
		return nil, ErrInvalidRequest
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	key := &model.ProviderKey{
		ProviderID:      providerID,
		Name:            req.Name,
		APIKeyEncrypted: req.APIKey,
		Status:          status,
		Weight:          defaultInt(req.Weight, 100),
		DailyLimit:      req.DailyLimit,
	}
	if err := s.repos.Providers.CreateKey(ctx, key); err != nil {
		return nil, err
	}
	return key, nil
}

func (s *adminCatalogService) DeleteProviderKey(ctx context.Context, id uint64) error {
	if id == 0 {
		return ErrInvalidRequest
	}
	return s.repos.Providers.DeleteKey(ctx, id)
}

func (s *adminCatalogService) ListImageModels(ctx context.Context, limit int, offset int) ([]model.ImageModel, error) {
	limit, offset = normalizePage(limit, offset)
	return s.repos.ImageModels.List(ctx, limit, offset)
}

func (s *adminCatalogService) CreateImageModel(ctx context.Context, req ImageModelRequest) (*model.ImageModel, error) {
	if req.Code == "" || req.DisplayName == "" || req.PriceCredits < 0 {
		return nil, ErrInvalidRequest
	}
	sizes, err := json.Marshal(req.SupportedSizes)
	if err != nil {
		return nil, err
	}
	imageModel := &model.ImageModel{
		Code:                req.Code,
		DisplayName:         req.DisplayName,
		Description:         req.Description,
		CoverURL:            req.CoverURL,
		PriceCredits:        req.PriceCredits,
		SupportedSizes:      datatypes.JSON(sizes),
		SupportTextToImage:  req.SupportTextToImage,
		SupportImageToImage: req.SupportImageToImage,
		SupportEdit:         req.SupportEdit,
		MaxImagesPerRequest: defaultInt(req.MaxImagesPerRequest, 1),
		AutoRefundOnFailure: req.AutoRefundOnFailure,
		Enabled:             req.Enabled,
		Recommended:         req.Recommended,
		SortOrder:           defaultInt(req.SortOrder, 100),
	}
	if err := s.repos.ImageModels.Create(ctx, imageModel); err != nil {
		return nil, err
	}
	return imageModel, nil
}

func (s *adminCatalogService) UpdateImageModel(ctx context.Context, id uint64, req ImageModelRequest) error {
	if id == 0 || req.Code == "" || req.DisplayName == "" || req.PriceCredits < 0 {
		return ErrInvalidRequest
	}
	sizes, err := json.Marshal(req.SupportedSizes)
	if err != nil {
		return err
	}
	return s.repos.ImageModels.Update(ctx, id, map[string]interface{}{
		"code":                   req.Code,
		"display_name":           req.DisplayName,
		"description":            req.Description,
		"cover_url":              req.CoverURL,
		"price_credits":          req.PriceCredits,
		"supported_sizes":        datatypes.JSON(sizes),
		"support_text_to_image":  req.SupportTextToImage,
		"support_image_to_image": req.SupportImageToImage,
		"support_edit":           req.SupportEdit,
		"max_images_per_request": defaultInt(req.MaxImagesPerRequest, 1),
		"auto_refund_on_failure": req.AutoRefundOnFailure,
		"enabled":                req.Enabled,
		"recommended":            req.Recommended,
		"sort_order":             defaultInt(req.SortOrder, 100),
	})
}

func (s *adminCatalogService) ListImageModelRoutes(ctx context.Context, modelID uint64) ([]model.ImageModelRoute, error) {
	return s.repos.ImageModels.ListRoutes(ctx, modelID)
}

func (s *adminCatalogService) CreateImageModelRoute(ctx context.Context, modelID uint64, req ImageModelRouteRequest) (*model.ImageModelRoute, error) {
	if modelID == 0 || req.ProviderID == 0 || req.ProviderModelName == "" {
		return nil, ErrInvalidRequest
	}
	extra, err := json.Marshal(req.ExtraConfig)
	if err != nil {
		return nil, err
	}
	route := &model.ImageModelRoute{
		ModelID:           modelID,
		ProviderID:        req.ProviderID,
		ProviderModelName: req.ProviderModelName,
		Enabled:           req.Enabled,
		Priority:          defaultInt(req.Priority, 100),
		Weight:            defaultInt(req.Weight, 100),
		ExtraConfig:       datatypes.JSON(extra),
	}
	if err := s.repos.ImageModels.CreateRoute(ctx, route); err != nil {
		return nil, err
	}
	return route, nil
}

func (s *adminCatalogService) UpdateImageModelRoute(ctx context.Context, id uint64, req ImageModelRouteRequest) error {
	if id == 0 || req.ProviderID == 0 || req.ProviderModelName == "" {
		return ErrInvalidRequest
	}
	extra, err := json.Marshal(req.ExtraConfig)
	if err != nil {
		return err
	}
	return s.repos.ImageModels.UpdateRoute(ctx, id, map[string]interface{}{
		"provider_id":         req.ProviderID,
		"provider_model_name": req.ProviderModelName,
		"enabled":             req.Enabled,
		"priority":            defaultInt(req.Priority, 100),
		"weight":              defaultInt(req.Weight, 100),
		"extra_config":        datatypes.JSON(extra),
	})
}

func defaultInt(value int, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}
