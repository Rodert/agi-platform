package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type channelRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Provider    string                 `json:"provider" binding:"required"`
	APIURL      string                 `json:"api_url" binding:"required"`
	APIKey      string                 `json:"api_key"`
	ExtraConfig map[string]interface{} `json:"extra_config"`
	IsActive    bool                   `json:"is_active"`
	Priority    int                    `json:"priority"`
}

type channelModelRequest struct {
	ModelName string `json:"model_name" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=image video text"`
	IsActive  *bool  `json:"is_active"`
}

type channelModelStatusRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *AdminConfigHandler) ListChannels(c *gin.Context) {
	rows, err := h.providerRepo.List()
	if err != nil {
		response.Error(c, err)
		return
	}
	for _, row := range rows {
		row.APIKey = "******"
	}
	response.Success(c, rows)
}

func (h *AdminConfigHandler) CreateChannel(c *gin.Context) {
	var req channelRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.APIKey == "" {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	extra, _ := json.Marshal(req.ExtraConfig)
	row := &model.AIProviderAccount{Name: req.Name, Provider: req.Provider, APIURL: req.APIURL, APIKey: req.APIKey, ExtraConfig: extra, IsActive: req.IsActive, Priority: req.Priority, HealthStatus: "unknown"}
	if row.Priority <= 0 {
		row.Priority = 100
	}
	if err := h.providerRepo.Create(row); err != nil {
		response.Error(c, err)
		return
	}
	row.APIKey = "******"
	response.Success(c, row)
}

func (h *AdminConfigHandler) UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	var req channelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	row, err := h.providerRepo.Find(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	extra, _ := json.Marshal(req.ExtraConfig)
	row.Name, row.Provider, row.APIURL, row.ExtraConfig = req.Name, req.Provider, req.APIURL, extra
	row.IsActive, row.Priority = req.IsActive, req.Priority
	if row.Priority <= 0 {
		row.Priority = 100
	}
	if req.APIKey != "" && req.APIKey != "******" {
		row.APIKey = req.APIKey
	}
	if err := h.providerRepo.Update(row); err != nil {
		response.Error(c, err)
		return
	}
	row.APIKey = "******"
	response.Success(c, row)
}

func (h *AdminConfigHandler) DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err == nil {
		err = h.providerRepo.Delete(id)
	}
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *AdminConfigHandler) SyncChannelModels(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	bindings, err := h.channelCatalogService.Sync(ctx, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, bindings)
}

func (h *AdminConfigHandler) BindChannelModel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	var req channelModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	binding, err := h.channelCatalogService.Bind(id, req.ModelName, req.Type, isActive)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, binding)
}

func (h *AdminConfigHandler) UpdateChannelModelStatus(c *gin.Context) {
	channelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	modelID, err := strconv.ParseInt(c.Param("modelID"), 10, 64)
	if err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	var req channelModelStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	if err := h.channelModelRepo.UpdateStatus(channelID, modelID, req.IsActive); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// Legacy endpoint names stay available while admin clients move to /channels.
func (h *AdminConfigHandler) ListProviderAccounts(c *gin.Context)  { h.ListChannels(c) }
func (h *AdminConfigHandler) CreateProviderAccount(c *gin.Context) { h.CreateChannel(c) }
func (h *AdminConfigHandler) UpdateProviderAccount(c *gin.Context) { h.UpdateChannel(c) }
func (h *AdminConfigHandler) DeleteProviderAccount(c *gin.Context) { h.DeleteChannel(c) }
