package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type PromptOptimizationHandler struct {
	service *service.PromptOptimizationService
}

func NewPromptOptimizationHandler(service *service.PromptOptimizationService) *PromptOptimizationHandler {
	return &PromptOptimizationHandler{service: service}
}

func (h *PromptOptimizationHandler) Optimize(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}
	var req dto.PromptOptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}
	result, err := h.service.Optimize(userID.(int64), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *PromptOptimizationHandler) ListAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	items, total, err := h.service.ListAdmin(page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page, pageSize)
}
