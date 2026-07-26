package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type WorkHandler struct {
	workService *service.WorkService
}

func NewWorkHandler(workService *service.WorkService) *WorkHandler {
	return &WorkHandler{
		workService: workService,
	}
}

// PublishWork 发布作品
// @Summary 发布作品
// @Tags 作品
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.PublishWorkRequest true "发布信息"
// @Success 200 {object} response.Response{data=dto.WorkResponse}
// @Router /api/v1/works [post]
func (h *WorkHandler) PublishWork(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req dto.PublishWorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	result, err := h.workService.PublishWork(userID.(int64), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "作品发布成功，等待审核", result)
}

// GetWork 获取作品详情
// @Summary 获取作品详情
// @Tags 作品
// @Accept json
// @Produce json
// @Param id path int true "作品ID"
// @Success 200 {object} response.Response{data=dto.WorkResponse}
// @Router /api/v1/works/{id} [get]
func (h *WorkHandler) GetWork(c *gin.Context) {
	workID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "作品ID无效", err.Error()))
		return
	}

	// 获取当前用户ID（可选）
	var userID int64
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(int64)
	}

	result, err := h.workService.GetWork(workID, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetWorkList 获取作品列表
// @Summary 获取作品列表
// @Tags 作品
// @Accept json
// @Produce json
// @Param category query string false "分类"
// @Param type query string false "类型"
// @Param user_id query int false "用户ID"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response{data=[]dto.WorkResponse}
// @Router /api/v1/works [get]
func (h *WorkHandler) GetWorkList(c *gin.Context) {
	var req dto.WorkListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	// 获取当前用户ID（可选）
	var userID int64
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(int64)
	}

	works, total, err := h.workService.GetWorkList(&req, userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, works, total, req.Page, req.PageSize)
}

// GetMyWorkList returns all works owned by the authenticated user, including
// records that are not visible on the public feed yet.
func (h *WorkHandler) GetMyWorkList(c *gin.Context) {
	var req dto.WorkListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	works, total, err := h.workService.GetMyWorkList(c.GetInt64("user_id"), req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, works, total, req.Page, req.PageSize)
}

// LikeWork 点赞作品
// @Summary 点赞作品
// @Tags 作品
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "作品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/works/{id}/like [post]
func (h *WorkHandler) LikeWork(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	workID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "作品ID无效", err.Error()))
		return
	}

	if err := h.workService.LikeWork(userID.(int64), workID); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "点赞成功", nil)
}

// UnlikeWork 取消点赞
// @Summary 取消点赞
// @Tags 作品
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "作品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/works/{id}/like [delete]
func (h *WorkHandler) UnlikeWork(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	workID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "作品ID无效", err.Error()))
		return
	}

	if err := h.workService.UnlikeWork(userID.(int64), workID); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "取消点赞成功", nil)
}

// CollectWork 收藏作品
// @Summary 收藏作品
// @Tags 作品
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "作品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/works/{id}/collect [post]
func (h *WorkHandler) CollectWork(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	workID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "作品ID无效", err.Error()))
		return
	}

	if err := h.workService.CollectWork(userID.(int64), workID); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "收藏成功", nil)
}

// UncollectWork 取消收藏
// @Summary 取消收藏
// @Tags 作品
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "作品ID"
// @Success 200 {object} response.Response
// @Router /api/v1/works/{id}/collect [delete]
func (h *WorkHandler) UncollectWork(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	workID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "作品ID无效", err.Error()))
		return
	}

	if err := h.workService.UncollectWork(userID.(int64), workID); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "取消收藏成功", nil)
}
