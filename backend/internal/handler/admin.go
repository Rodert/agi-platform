package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	result, err := h.adminService.Login(&req, c.ClientIP())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetStats 获取统计数据
func (h *AdminHandler) GetStats(c *gin.Context) {
	result, err := h.adminService.GetStats()
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetPendingWorks 获取待审核作品列表
func (h *AdminHandler) GetPendingWorks(c *gin.Context) {
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

	works, total, err := h.adminService.GetPendingWorks(req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, works, total, req.Page, req.PageSize)
}

// AuditWork 审核作品
func (h *AdminHandler) AuditWork(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeBadRequest, "无效的作品ID"))
		return
	}

	var req dto.AdminWorkAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	adminID, exists := c.Get("admin_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}
	err = h.adminService.AuditWork(adminID.(int64), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// GetUserList 获取用户列表
func (h *AdminHandler) GetUserList(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	result, err := h.adminService.GetUserList(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// CreateUser 创建用户
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	err := h.adminService.CreateUser(&req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.Error(c, errors.New(errors.ErrCodeBadRequest, "无效的用户ID")); return }
	var req dto.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	if err := h.adminService.UpdateUser(id, &req); err != nil { response.Error(c, err); return }
	response.Success(c, nil)
}

// UpdateUserStatus 更新用户状态
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeBadRequest, "无效的用户ID"))
		return
	}

	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	err = h.adminService.UpdateUserStatus(id, req.IsActive)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
