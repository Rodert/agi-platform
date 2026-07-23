package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type AnnouncementHandler struct{ service *service.AnnouncementService }

func NewAnnouncementHandler(service *service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{service: service}
}

func (h *AnnouncementHandler) ListPublished(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, total, err := h.service.ListPublished(page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page, pageSize)
}

func (h *AnnouncementHandler) ListAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, total, err := h.service.ListAdmin(page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page, pageSize)
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	var input service.AnnouncementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}
	item, err := h.service.Create(&input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *AnnouncementHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeBadRequest, "通知ID无效"))
		return
	}
	var input service.AnnouncementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}
	item, err := h.service.Update(id, &input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, item)
}

func (h *AnnouncementHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeBadRequest, "通知ID无效"))
		return
	}
	if err := h.service.Delete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}
