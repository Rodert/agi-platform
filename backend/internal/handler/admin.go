package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type AdminHandler struct {
	adminService *service.AdminService
	redeemCodeService *service.RedeemCodeService
}

// GetReport returns operational reporting data for an inclusive date range.
func (h *AdminHandler) GetReport(c *gin.Context) {
	var req dto.AdminReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "日期参数错误", err.Error()))
		return
	}
	start, err := time.ParseInLocation("2006-01-02", req.StartDate, time.Local)
	if err != nil { response.Error(c, errors.New(errors.ErrCodeBadRequest, "开始日期无效")); return }
	end, err := time.ParseInLocation("2006-01-02", req.EndDate, time.Local)
	if err != nil { response.Error(c, errors.New(errors.ErrCodeBadRequest, "结束日期无效")); return }
	if end.Before(start) || end.Sub(start) > 366*24*time.Hour {
		response.Error(c, errors.New(errors.ErrCodeBadRequest, "日期范围需在一年内且结束日期不早于开始日期"))
		return
	}
	result, err := h.adminService.GetReport(start, end.AddDate(0, 0, 1))
	if err != nil { response.Error(c, err); return }
	response.Success(c, result)
}

func NewAdminHandler(adminService *service.AdminService, redeemCodeService *service.RedeemCodeService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		redeemCodeService: redeemCodeService,
	}
}

func (h *AdminHandler) CreateRedeemCodes(c *gin.Context) {
	var req dto.AdminCreateRedeemCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	items, err := h.redeemCodeService.Create(c.GetInt64("admin_id"), &req)
	if err != nil { response.Error(c, err); return }
	response.Success(c, items)
}

func (h *AdminHandler) ListRedeemCodes(c *gin.Context) {
	var req dto.AdminRedeemCodeListRequest
	if err := c.ShouldBindQuery(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	if req.Page < 1 { req.Page = 1 }; if req.PageSize < 1 { req.PageSize = 20 }
	items, total, err := h.redeemCodeService.List(&req)
	if err != nil { response.Error(c, err); return }
	response.Page(c, items, total, req.Page, req.PageSize)
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

func (h *AdminHandler) GetProfile(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists { response.Error(c, errors.ErrUnauthorized); return }
	profile, err := h.adminService.GetProfile(adminID.(int64))
	if err != nil { response.Error(c, err); return }
	response.Success(c, profile)
}

func (h *AdminHandler) UpdateProfile(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists { response.Error(c, errors.ErrUnauthorized); return }
	var req dto.AdminUpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	profile, err := h.adminService.UpdateProfile(adminID.(int64), &req)
	if err != nil { response.Error(c, err); return }
	response.Success(c, profile)
}

func (h *AdminHandler) ListAdmins(c *gin.Context) {
	items, err := h.adminService.ListAdmins()
	if err != nil { response.Error(c, err); return }
	response.Success(c, items)
}

func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req dto.CreateAdminManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	item, err := h.adminService.CreateAdmin(c.GetInt64("admin_id"), &req)
	if err != nil { response.Error(c, err); return }
	response.Success(c, item)
}

func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.Error(c, errors.New(errors.ErrCodeBadRequest, "无效的管理员ID")); return }
	var req dto.UpdateAdminManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	item, err := h.adminService.UpdateAdmin(c.GetInt64("admin_id"), id, &req)
	if err != nil { response.Error(c, err); return }
	response.Success(c, item)
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

// GetLogs returns persisted administrative audit logs.
func (h *AdminHandler) GetLogs(c *gin.Context) {
	var req dto.AdminLogListRequest
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
	logs, total, err := h.adminService.GetLogs(&req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, logs, total, req.Page, req.PageSize)
}

func (h *AdminHandler) ListDatabaseTables(c *gin.Context) {
	tables, err := h.adminService.ListDatabaseTables()
	if err != nil { response.Error(c, err); return }
	response.Success(c, tables)
}

func (h *AdminHandler) GetDatabaseTable(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 { page = 1 }
	if pageSize < 1 { pageSize = 20 }
	data, err := h.adminService.GetDatabaseTable(c.GetInt64("admin_id"), c.Param("table"), page, pageSize)
	if err != nil { response.Error(c, err); return }
	response.Success(c, data)
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

// GetWorks lists all work records so reviewed items remain manageable after review.
func (h *AdminHandler) GetWorks(c *gin.Context) {
	var req dto.AdminWorkListRequest
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
	works, total, err := h.adminService.GetWorks(req.Status, req.Page, req.PageSize)
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

// UpdateWorkPublicationStatus takes an approved work off the public feed or republishes it.
func (h *AdminHandler) UpdateWorkPublicationStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeBadRequest, "无效的作品ID"))
		return
	}
	var req dto.AdminWorkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}
	adminID, exists := c.Get("admin_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}
	if err := h.adminService.UpdateWorkPublicationStatus(adminID.(int64), id, &req); err != nil {
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

func (h *AdminHandler) GetTaskList(c *gin.Context) {
	var req dto.AdminTaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	tasks, total, err := h.adminService.GetTaskList(&req)
	if err != nil { response.Error(c, err); return }
	response.Page(c, tasks, total, req.Page, req.PageSize)
}

// CreateUser 创建用户
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	adminID, exists := c.Get("admin_id")
	if !exists { response.Error(c, errors.ErrUnauthorized); return }
	err := h.adminService.CreateUser(adminID.(int64), &req)
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
	adminID, exists := c.Get("admin_id")
	if !exists { response.Error(c, errors.ErrUnauthorized); return }
	if err := h.adminService.UpdateUser(adminID.(int64), id, &req); err != nil { response.Error(c, err); return }
	response.Success(c, nil)
}

func (h *AdminHandler) RechargeUserCredit(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.Error(c, errors.New(errors.ErrCodeBadRequest, "无效的用户ID")); return }
	adminID, exists := c.Get("admin_id")
	if !exists { response.Error(c, errors.ErrUnauthorized); return }
	var req dto.AdminRechargeCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	result, err := h.adminService.RechargeUserCredit(adminID.(int64), userID, &req)
	if err != nil { response.Error(c, err); return }
	response.Success(c, result)
}

// GetUserCreditLedgers returns one user's immutable credit ledger.
func (h *AdminHandler) GetUserCreditLedgers(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.Error(c, errors.New(errors.ErrCodeBadRequest, "无效的用户ID")); return }
	var req dto.AdminCreditLedgerListRequest
	if err := c.ShouldBindQuery(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	if req.Page < 1 { req.Page = 1 }
	if req.PageSize < 1 { req.PageSize = 20 }
	items, total, err := h.adminService.GetUserCreditLedgers(userID, &req)
	if err != nil { response.Error(c, err); return }
	response.Page(c, items, total, req.Page, req.PageSize)
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
