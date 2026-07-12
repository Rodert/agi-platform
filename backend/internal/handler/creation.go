package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type CreationHandler struct {
	creationService *service.CreationService
}

func NewCreationHandler(creationService *service.CreationService) *CreationHandler {
	return &CreationHandler{
		creationService: creationService,
	}
}

// CreateImageTask 创建图片生成任务
// @Summary 创建图片生成任务
// @Tags 创作
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateImageTaskRequest true "任务信息"
// @Success 200 {object} response.Response{data=dto.TaskResponse}
// @Router /api/v1/generation/image [post]
func (h *CreationHandler) CreateImageTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req dto.CreateImageTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	result, err := h.creationService.CreateImageTask(userID.(int64), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "任务创建成功", result)
}

// CreateVideoTask 创建视频生成任务
// @Summary 创建视频生成任务
// @Tags 创作
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateVideoTaskRequest true "任务信息"
// @Success 200 {object} response.Response{data=dto.TaskResponse}
// @Router /api/v1/generation/video [post]
func (h *CreationHandler) CreateVideoTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req dto.CreateVideoTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	result, err := h.creationService.CreateVideoTask(userID.(int64), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "任务创建成功", result)
}

// GetTask 获取任务详情
// @Summary 获取任务详情
// @Tags 创作
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response{data=dto.TaskResponse}
// @Router /api/v1/tasks/{id} [get]
func (h *CreationHandler) GetTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "任务ID无效", err.Error()))
		return
	}

	result, err := h.creationService.GetTask(userID.(int64), taskID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetTaskList 获取任务列表
// @Summary 获取任务列表
// @Tags 创作
// @Accept json
// @Produce json
// @Security Bearer
// @Param status query string false "状态筛选"
// @Param type query string false "类型筛选"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response{data=[]dto.TaskResponse}
// @Router /api/v1/tasks [get]
func (h *CreationHandler) GetTaskList(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req dto.TaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	tasks, total, err := h.creationService.GetTaskList(userID.(int64), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, tasks, total, req.Page, req.PageSize)
}

// GetModels 获取模型列表
// @Summary 获取模型列表
// @Tags 创作
// @Accept json
// @Produce json
// @Param type query string false "类型筛选: image/video"
// @Success 200 {object} response.Response{data=[]dto.GetModelsResponse}
// @Router /api/v1/generation/models [get]
func (h *CreationHandler) GetModels(c *gin.Context) {
	modelType := c.Query("type")

	models, err := h.creationService.GetModels(modelType)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, models)
}
