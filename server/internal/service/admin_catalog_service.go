package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
	DeleteImageModel(ctx context.Context, id uint64) error
	ListImageModelRoutes(ctx context.Context, modelID uint64) ([]model.ImageModelRoute, error)
	CreateImageModelRoute(ctx context.Context, modelID uint64, req ImageModelRouteRequest) (*model.ImageModelRoute, error)
	UpdateImageModelRoute(ctx context.Context, id uint64, req ImageModelRouteRequest) error
	ListVideoModels(ctx context.Context, limit int, offset int) ([]model.VideoModel, error)
	CreateVideoModel(ctx context.Context, req VideoModelRequest) (*model.VideoModel, error)
	UpdateVideoModel(ctx context.Context, id uint64, req VideoModelRequest) error
	DeleteVideoModel(ctx context.Context, id uint64) error
	ListVideoModelRoutes(ctx context.Context, modelID uint64) ([]model.VideoModelRoute, error)
	CreateVideoModelRoute(ctx context.Context, modelID uint64, req VideoModelRouteRequest) (*model.VideoModelRoute, error)
	UpdateVideoModelRoute(ctx context.Context, id uint64, req VideoModelRouteRequest) error
	SaveUpstreamIntegration(ctx context.Context, req UpstreamIntegrationRequest) (*UpstreamIntegrationResult, error)
	QueryUpstreamModels(ctx context.Context, req QueryUpstreamModelsRequest) ([]UpstreamModel, error)
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
	ProviderKeyID     *uint64
	ProviderModelName string
	Enabled           bool
	Priority          int
	Weight            int
	ExtraConfig       map[string]interface{}
}

type VideoModelRequest struct {
	Code                  string
	DisplayName           string
	Description           string
	PriceCredits          int64
	SupportedAspectRatios []string
	SupportedSeconds      []int
	Enabled               bool
	Recommended           bool
	SortOrder             int
}

type VideoModelRouteRequest struct {
	ProviderID        uint64
	ProviderKeyID     *uint64
	ProviderModelName string
	Enabled           bool
	Priority          int
	Weight            int
	ExtraConfig       map[string]interface{}
}

type QueryUpstreamModelsRequest struct {
	BaseURL string
	APIKey  string
}

type UpstreamIntegrationRequest struct {
	Provider ProviderRequest
	APIKey   *ProviderKeyRequest
	Models   []UpstreamIntegrationModelRequest
}

type UpstreamIntegrationModelRequest struct {
	ModelType         string
	ProviderKeyID     *uint64
	ProviderModelName string
	Enabled           bool
	Priority          int
	Weight            int
	ExtraConfig       map[string]interface{}
	ImageModel        ImageModelRequest
	VideoModel        VideoModelRequest
}

type UpstreamIntegrationResult struct {
	Provider    *model.Provider `json:"provider"`
	KeyCreated  bool            `json:"key_created"`
	ImageModels int             `json:"image_models"`
	VideoModels int             `json:"video_models"`
}

type UpstreamModel struct {
	ID     string `json:"id"`
	Object string `json:"object,omitempty"`
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

func (s *adminCatalogService) DeleteImageModel(ctx context.Context, id uint64) error {
	if id == 0 {
		return ErrInvalidRequest
	}
	count, err := s.repos.ImageTasks.CountByModelID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: image model has tasks, disable it instead", ErrInvalidRequest)
	}
	return s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		if err := s.repos.ImageModels.DeleteRoutesByModelID(ctx, tx, id); err != nil {
			return err
		}
		return s.repos.ImageModels.Delete(ctx, tx, id)
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
		ProviderKeyID:     req.ProviderKeyID,
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
		"provider_key_id":     req.ProviderKeyID,
		"provider_model_name": req.ProviderModelName,
		"enabled":             req.Enabled,
		"priority":            defaultInt(req.Priority, 100),
		"weight":              defaultInt(req.Weight, 100),
		"extra_config":        datatypes.JSON(extra),
	})
}

func (s *adminCatalogService) ListVideoModels(ctx context.Context, limit int, offset int) ([]model.VideoModel, error) {
	limit, offset = normalizePage(limit, offset)
	return s.repos.Videos.ListAllModels(ctx, limit, offset)
}

func (s *adminCatalogService) CreateVideoModel(ctx context.Context, req VideoModelRequest) (*model.VideoModel, error) {
	if req.Code == "" || req.DisplayName == "" || req.PriceCredits < 0 {
		return nil, ErrInvalidRequest
	}
	ratios, err := json.Marshal(req.SupportedAspectRatios)
	if err != nil {
		return nil, err
	}
	seconds, err := json.Marshal(req.SupportedSeconds)
	if err != nil {
		return nil, err
	}
	videoModel := &model.VideoModel{
		Code:                  req.Code,
		DisplayName:           req.DisplayName,
		Description:           req.Description,
		PriceCredits:          req.PriceCredits,
		SupportedAspectRatios: datatypes.JSON(ratios),
		SupportedSeconds:      datatypes.JSON(seconds),
		Enabled:               req.Enabled,
		Recommended:           req.Recommended,
		SortOrder:             defaultInt(req.SortOrder, 100),
	}
	if err := s.repos.Videos.CreateModel(ctx, videoModel); err != nil {
		return nil, err
	}
	return videoModel, nil
}

func (s *adminCatalogService) UpdateVideoModel(ctx context.Context, id uint64, req VideoModelRequest) error {
	if id == 0 || req.Code == "" || req.DisplayName == "" || req.PriceCredits < 0 {
		return ErrInvalidRequest
	}
	ratios, err := json.Marshal(req.SupportedAspectRatios)
	if err != nil {
		return err
	}
	seconds, err := json.Marshal(req.SupportedSeconds)
	if err != nil {
		return err
	}
	return s.repos.Videos.UpdateModel(ctx, id, map[string]interface{}{
		"code":                    req.Code,
		"display_name":            req.DisplayName,
		"description":             req.Description,
		"price_credits":           req.PriceCredits,
		"supported_aspect_ratios": datatypes.JSON(ratios),
		"supported_seconds":       datatypes.JSON(seconds),
		"enabled":                 req.Enabled,
		"recommended":             req.Recommended,
		"sort_order":              defaultInt(req.SortOrder, 100),
	})
}

func (s *adminCatalogService) DeleteVideoModel(ctx context.Context, id uint64) error {
	if id == 0 {
		return ErrInvalidRequest
	}
	count, err := s.repos.Videos.CountTasksByModelID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: video model has tasks, disable it instead", ErrInvalidRequest)
	}
	return s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		if err := s.repos.Videos.DeleteRoutesByModelID(ctx, tx, id); err != nil {
			return err
		}
		return s.repos.Videos.DeleteModel(ctx, tx, id)
	})
}

func (s *adminCatalogService) ListVideoModelRoutes(ctx context.Context, modelID uint64) ([]model.VideoModelRoute, error) {
	return s.repos.Videos.ListRoutes(ctx, modelID)
}

func (s *adminCatalogService) CreateVideoModelRoute(ctx context.Context, modelID uint64, req VideoModelRouteRequest) (*model.VideoModelRoute, error) {
	if modelID == 0 || req.ProviderID == 0 || req.ProviderModelName == "" {
		return nil, ErrInvalidRequest
	}
	extra, err := json.Marshal(req.ExtraConfig)
	if err != nil {
		return nil, err
	}
	route := &model.VideoModelRoute{
		ModelID:           modelID,
		ProviderID:        req.ProviderID,
		ProviderKeyID:     req.ProviderKeyID,
		ProviderModelName: req.ProviderModelName,
		Enabled:           req.Enabled,
		Priority:          defaultInt(req.Priority, 100),
		Weight:            defaultInt(req.Weight, 100),
		ExtraConfig:       datatypes.JSON(extra),
	}
	if err := s.repos.Videos.CreateRoute(ctx, route); err != nil {
		return nil, err
	}
	return route, nil
}

func (s *adminCatalogService) UpdateVideoModelRoute(ctx context.Context, id uint64, req VideoModelRouteRequest) error {
	if id == 0 || req.ProviderID == 0 || req.ProviderModelName == "" {
		return ErrInvalidRequest
	}
	extra, err := json.Marshal(req.ExtraConfig)
	if err != nil {
		return err
	}
	return s.repos.Videos.UpdateRoute(ctx, id, map[string]interface{}{
		"provider_id":         req.ProviderID,
		"provider_key_id":     req.ProviderKeyID,
		"provider_model_name": req.ProviderModelName,
		"enabled":             req.Enabled,
		"priority":            defaultInt(req.Priority, 100),
		"weight":              defaultInt(req.Weight, 100),
		"extra_config":        datatypes.JSON(extra),
	})
}

func (s *adminCatalogService) SaveUpstreamIntegration(ctx context.Context, req UpstreamIntegrationRequest) (*UpstreamIntegrationResult, error) {
	providerReq := req.Provider
	providerReq.Code = strings.TrimSpace(providerReq.Code)
	providerReq.Name = strings.TrimSpace(providerReq.Name)
	providerReq.Type = strings.TrimSpace(providerReq.Type)
	providerReq.BaseURL = strings.TrimSpace(providerReq.BaseURL)
	if providerReq.Code == "" || providerReq.Name == "" || providerReq.Type == "" || providerReq.BaseURL == "" || len(req.Models) == 0 {
		return nil, ErrInvalidRequest
	}

	provider, err := s.upsertProvider(ctx, providerReq)
	if err != nil {
		return nil, err
	}

	var keyID *uint64
	keyCreated := false
	if req.APIKey != nil && strings.TrimSpace(req.APIKey.APIKey) != "" {
		keyReq := *req.APIKey
		keyReq.Name = strings.TrimSpace(keyReq.Name)
		keyReq.APIKey = strings.TrimSpace(keyReq.APIKey)
		key, err := s.CreateProviderKey(ctx, provider.ID, keyReq)
		if err != nil {
			return nil, err
		}
		keyID = &key.ID
		keyCreated = true
	} else if key, err := s.repos.Providers.PickActiveKey(ctx, provider.ID); err == nil {
		keyID = &key.ID
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	result := &UpstreamIntegrationResult{Provider: provider, KeyCreated: keyCreated}
	for _, item := range req.Models {
		item.ProviderModelName = strings.TrimSpace(item.ProviderModelName)
		item.ModelType = strings.TrimSpace(item.ModelType)
		if item.ProviderKeyID == nil {
			item.ProviderKeyID = keyID
		}
		if item.Weight == 0 {
			item.Weight = 100
		}
		if item.ModelType == "video" {
			if err := s.saveVideoIntegrationModel(ctx, provider.ID, item); err != nil {
				return nil, err
			}
			result.VideoModels++
			continue
		}
		if item.ModelType != "" && item.ModelType != "image" {
			return nil, ErrInvalidRequest
		}
		if err := s.saveImageIntegrationModel(ctx, provider.ID, item); err != nil {
			return nil, err
		}
		result.ImageModels++
	}
	return result, nil
}

func (s *adminCatalogService) upsertProvider(ctx context.Context, req ProviderRequest) (*model.Provider, error) {
	provider, err := s.repos.Providers.FindByCode(ctx, req.Code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return s.CreateProvider(ctx, req)
		}
		return nil, err
	}
	if provider != nil {
		if err := s.UpdateProvider(ctx, provider.ID, req); err != nil {
			return nil, err
		}
		updated := *provider
		updated.Code = req.Code
		updated.Name = req.Name
		updated.Type = req.Type
		updated.BaseURL = req.BaseURL
		updated.Enabled = req.Enabled
		updated.TimeoutSeconds = defaultInt(req.TimeoutSeconds, 60)
		updated.RetryCount = req.RetryCount
		updated.Priority = defaultInt(req.Priority, 100)
		updated.DailyLimit = req.DailyLimit
		updated.Remark = req.Remark
		return &updated, nil
	}
	return s.CreateProvider(ctx, req)
}

func (s *adminCatalogService) saveImageIntegrationModel(ctx context.Context, providerID uint64, item UpstreamIntegrationModelRequest) error {
	req := item.ImageModel
	req.Code = strings.TrimSpace(req.Code)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Description = strings.TrimSpace(req.Description)
	if req.Code == "" || req.DisplayName == "" || item.ProviderModelName == "" {
		return ErrInvalidRequest
	}
	imageModel, err := s.findImageModelByCode(ctx, req.Code)
	if err != nil {
		return err
	}
	if imageModel == nil {
		imageModel, err = s.CreateImageModel(ctx, req)
		if err != nil {
			return err
		}
	} else if err := s.UpdateImageModel(ctx, imageModel.ID, req); err != nil {
		return err
	}
	return s.upsertImageModelRoute(ctx, imageModel.ID, providerID, ImageModelRouteRequest{
		ProviderID:        providerID,
		ProviderKeyID:     item.ProviderKeyID,
		ProviderModelName: item.ProviderModelName,
		Enabled:           item.Enabled,
		Priority:          item.Priority,
		Weight:            item.Weight,
		ExtraConfig:       item.ExtraConfig,
	})
}

func (s *adminCatalogService) saveVideoIntegrationModel(ctx context.Context, providerID uint64, item UpstreamIntegrationModelRequest) error {
	req := item.VideoModel
	req.Code = strings.TrimSpace(req.Code)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.Description = strings.TrimSpace(req.Description)
	if req.Code == "" || req.DisplayName == "" || item.ProviderModelName == "" {
		return ErrInvalidRequest
	}
	videoModel, err := s.findVideoModelByCode(ctx, req.Code)
	if err != nil {
		return err
	}
	if videoModel == nil {
		videoModel, err = s.CreateVideoModel(ctx, req)
		if err != nil {
			return err
		}
	} else if err := s.UpdateVideoModel(ctx, videoModel.ID, req); err != nil {
		return err
	}
	return s.upsertVideoModelRoute(ctx, videoModel.ID, providerID, VideoModelRouteRequest{
		ProviderID:        providerID,
		ProviderKeyID:     item.ProviderKeyID,
		ProviderModelName: item.ProviderModelName,
		Enabled:           item.Enabled,
		Priority:          item.Priority,
		Weight:            item.Weight,
		ExtraConfig:       item.ExtraConfig,
	})
}

func (s *adminCatalogService) findImageModelByCode(ctx context.Context, code string) (*model.ImageModel, error) {
	imageModel, err := s.repos.ImageModels.FindAnyByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return imageModel, nil
}

func (s *adminCatalogService) findVideoModelByCode(ctx context.Context, code string) (*model.VideoModel, error) {
	videoModel, err := s.repos.Videos.FindAnyModelByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return videoModel, nil
}

func (s *adminCatalogService) upsertImageModelRoute(ctx context.Context, modelID uint64, providerID uint64, req ImageModelRouteRequest) error {
	routes, err := s.repos.ImageModels.ListRoutes(ctx, modelID)
	if err != nil {
		return err
	}
	for _, route := range routes {
		if route.ProviderID == providerID {
			return s.UpdateImageModelRoute(ctx, route.ID, req)
		}
	}
	_, err = s.CreateImageModelRoute(ctx, modelID, req)
	return err
}

func (s *adminCatalogService) upsertVideoModelRoute(ctx context.Context, modelID uint64, providerID uint64, req VideoModelRouteRequest) error {
	routes, err := s.repos.Videos.ListRoutes(ctx, modelID)
	if err != nil {
		return err
	}
	for _, route := range routes {
		if route.ProviderID == providerID {
			return s.UpdateVideoModelRoute(ctx, route.ID, req)
		}
	}
	_, err = s.CreateVideoModelRoute(ctx, modelID, req)
	return err
}

func (s *adminCatalogService) QueryUpstreamModels(ctx context.Context, req QueryUpstreamModelsRequest) ([]UpstreamModel, error) {
	baseURL := strings.TrimSpace(req.BaseURL)
	apiKey := strings.TrimSpace(req.APIKey)
	if baseURL == "" || apiKey == "" {
		return nil, ErrInvalidRequest
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamModelsEndpoint(baseURL), nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream models request failed: status=%d body=%s", resp.StatusCode, truncateForAdmin(raw, 800))
	}

	var payload struct {
		Data []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	if payload.Data == nil {
		return nil, errors.New("upstream response did not include data")
	}

	models := make([]UpstreamModel, 0, len(payload.Data))
	seen := map[string]struct{}{}
	for _, item := range payload.Data {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		models = append(models, UpstreamModel{ID: id, Object: item.Object})
	}
	return models, nil
}

func upstreamModelsEndpoint(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(trimmed, "/models") {
		return trimmed
	}
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed + "/models"
	}
	return trimmed + "/v1/models"
}

func truncateForAdmin(raw []byte, limit int) string {
	value := string(raw)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}

func defaultInt(value int, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}
