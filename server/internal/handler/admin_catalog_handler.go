package handler

import (
	"net/http"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminCatalogHandler interface {
	ListProviders(c *gin.Context)
	CreateProvider(c *gin.Context)
	UpdateProvider(c *gin.Context)
	ListProviderKeys(c *gin.Context)
	CreateProviderKey(c *gin.Context)
	DeleteProviderKey(c *gin.Context)
	ListImageModels(c *gin.Context)
	CreateImageModel(c *gin.Context)
	UpdateImageModel(c *gin.Context)
	DeleteImageModel(c *gin.Context)
	ListImageModelRoutes(c *gin.Context)
	CreateImageModelRoute(c *gin.Context)
	UpdateImageModelRoute(c *gin.Context)
	ListVideoModels(c *gin.Context)
	CreateVideoModel(c *gin.Context)
	UpdateVideoModel(c *gin.Context)
	DeleteVideoModel(c *gin.Context)
	ListVideoModelRoutes(c *gin.Context)
	CreateVideoModelRoute(c *gin.Context)
	UpdateVideoModelRoute(c *gin.Context)
	QueryUpstreamModels(c *gin.Context)
}

type adminCatalogHandler struct {
	service service.AdminCatalogService
}

type providerRequest struct {
	Code           string `json:"code" binding:"required"`
	Name           string `json:"name" binding:"required"`
	Type           string `json:"type" binding:"required"`
	BaseURL        string `json:"base_url"`
	Enabled        bool   `json:"enabled"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	RetryCount     int    `json:"retry_count"`
	Priority       int    `json:"priority"`
	DailyLimit     *int64 `json:"daily_limit"`
	Remark         string `json:"remark"`
}

type providerKeyRequest struct {
	Name       string `json:"name"`
	APIKey     string `json:"api_key" binding:"required"`
	Status     string `json:"status"`
	Weight     int    `json:"weight"`
	DailyLimit *int64 `json:"daily_limit"`
}

type imageModelRequest struct {
	Code                string   `json:"code" binding:"required"`
	DisplayName         string   `json:"display_name" binding:"required"`
	Description         string   `json:"description"`
	CoverURL            string   `json:"cover_url"`
	PriceCredits        int64    `json:"price_credits"`
	SupportedSizes      []string `json:"supported_sizes"`
	SupportTextToImage  bool     `json:"support_text_to_image"`
	SupportImageToImage bool     `json:"support_image_to_image"`
	SupportEdit         bool     `json:"support_edit"`
	MaxImagesPerRequest int      `json:"max_images_per_request"`
	AutoRefundOnFailure bool     `json:"auto_refund_on_failure"`
	Enabled             bool     `json:"enabled"`
	Recommended         bool     `json:"recommended"`
	SortOrder           int      `json:"sort_order"`
}

type imageModelRouteRequest struct {
	ProviderID        uint64                 `json:"provider_id" binding:"required"`
	ProviderKeyID     *uint64                `json:"provider_key_id"`
	ProviderModelName string                 `json:"provider_model_name" binding:"required"`
	Enabled           bool                   `json:"enabled"`
	Priority          int                    `json:"priority"`
	Weight            int                    `json:"weight"`
	ExtraConfig       map[string]interface{} `json:"extra_config"`
}

type videoModelRequest struct {
	Code                  string   `json:"code" binding:"required"`
	DisplayName           string   `json:"display_name" binding:"required"`
	Description           string   `json:"description"`
	PriceCredits          int64    `json:"price_credits"`
	SupportedAspectRatios []string `json:"supported_aspect_ratios"`
	SupportedSeconds      []int    `json:"supported_seconds"`
	Enabled               bool     `json:"enabled"`
	Recommended           bool     `json:"recommended"`
	SortOrder             int      `json:"sort_order"`
}

type videoModelRouteRequest struct {
	ProviderID        uint64                 `json:"provider_id" binding:"required"`
	ProviderKeyID     *uint64                `json:"provider_key_id"`
	ProviderModelName string                 `json:"provider_model_name" binding:"required"`
	Enabled           bool                   `json:"enabled"`
	Priority          int                    `json:"priority"`
	Weight            int                    `json:"weight"`
	ExtraConfig       map[string]interface{} `json:"extra_config"`
}

type queryUpstreamModelsRequest struct {
	BaseURL string `json:"base_url" binding:"required"`
	APIKey  string `json:"api_key" binding:"required"`
}

func NewAdminCatalogHandler(service service.AdminCatalogService) AdminCatalogHandler {
	return &adminCatalogHandler{service: service}
}

func (h *adminCatalogHandler) ListProviders(c *gin.Context) {
	items, err := h.service.ListProviders(c.Request.Context(), queryInt(c, "limit", 20), queryInt(c, "offset", 0))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *adminCatalogHandler) CreateProvider(c *gin.Context) {
	var req providerRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.service.CreateProvider(c.Request.Context(), service.ProviderRequest(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *adminCatalogHandler) UpdateProvider(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req providerRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.service.UpdateProvider(c.Request.Context(), id, service.ProviderRequest(req)); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"updated": true})
}

func (h *adminCatalogHandler) ListProviderKeys(c *gin.Context) {
	providerID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	items, err := h.service.ListProviderKeys(c.Request.Context(), providerID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *adminCatalogHandler) CreateProviderKey(c *gin.Context) {
	providerID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req providerKeyRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.service.CreateProviderKey(c.Request.Context(), providerID, service.ProviderKeyRequest(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *adminCatalogHandler) DeleteProviderKey(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteProviderKey(c.Request.Context(), id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *adminCatalogHandler) ListImageModels(c *gin.Context) {
	items, err := h.service.ListImageModels(c.Request.Context(), queryInt(c, "limit", 20), queryInt(c, "offset", 0))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *adminCatalogHandler) CreateImageModel(c *gin.Context) {
	var req imageModelRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.service.CreateImageModel(c.Request.Context(), service.ImageModelRequest(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *adminCatalogHandler) UpdateImageModel(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req imageModelRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.service.UpdateImageModel(c.Request.Context(), id, service.ImageModelRequest(req)); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"updated": true})
}

func (h *adminCatalogHandler) DeleteImageModel(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteImageModel(c.Request.Context(), id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *adminCatalogHandler) ListImageModelRoutes(c *gin.Context) {
	modelID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	items, err := h.service.ListImageModelRoutes(c.Request.Context(), modelID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *adminCatalogHandler) CreateImageModelRoute(c *gin.Context) {
	modelID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req imageModelRouteRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.service.CreateImageModelRoute(c.Request.Context(), modelID, service.ImageModelRouteRequest(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *adminCatalogHandler) UpdateImageModelRoute(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req imageModelRouteRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.service.UpdateImageModelRoute(c.Request.Context(), id, service.ImageModelRouteRequest(req)); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"updated": true})
}

func (h *adminCatalogHandler) ListVideoModels(c *gin.Context) {
	items, err := h.service.ListVideoModels(c.Request.Context(), queryInt(c, "limit", 20), queryInt(c, "offset", 0))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *adminCatalogHandler) CreateVideoModel(c *gin.Context) {
	var req videoModelRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.service.CreateVideoModel(c.Request.Context(), service.VideoModelRequest(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *adminCatalogHandler) UpdateVideoModel(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req videoModelRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.service.UpdateVideoModel(c.Request.Context(), id, service.VideoModelRequest(req)); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"updated": true})
}

func (h *adminCatalogHandler) DeleteVideoModel(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteVideoModel(c.Request.Context(), id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *adminCatalogHandler) ListVideoModelRoutes(c *gin.Context) {
	modelID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	items, err := h.service.ListVideoModelRoutes(c.Request.Context(), modelID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *adminCatalogHandler) CreateVideoModelRoute(c *gin.Context) {
	modelID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req videoModelRouteRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.service.CreateVideoModelRoute(c.Request.Context(), modelID, service.VideoModelRouteRequest(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *adminCatalogHandler) UpdateVideoModelRoute(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req videoModelRouteRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.service.UpdateVideoModelRoute(c.Request.Context(), id, service.VideoModelRouteRequest(req)); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"updated": true})
}

func (h *adminCatalogHandler) QueryUpstreamModels(c *gin.Context) {
	var req queryUpstreamModelsRequest
	if !bindJSON(c, &req) {
		return
	}
	items, err := h.service.QueryUpstreamModels(c.Request.Context(), service.QueryUpstreamModelsRequest(req))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, items)
}

func bindJSON(c *gin.Context, dest interface{}) bool {
	if err := c.ShouldBindJSON(dest); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}
