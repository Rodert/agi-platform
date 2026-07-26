package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type UserHandler struct {
	userService *service.UserService
	redeemCodeService *service.RedeemCodeService
}

func (h *UserHandler) GetCreditPackages(c *gin.Context) {
	items, err := h.redeemCodeService.ListActivePackages()
	if err != nil { response.Error(c, err); return }
	response.Success(c, items)
}

func NewUserHandler(userService *service.UserService, redeemCodeService *service.RedeemCodeService) *UserHandler {
	return &UserHandler{
		userService: userService,
		redeemCodeService: redeemCodeService,
	}
}

func (h *UserHandler) RedeemCode(c *gin.Context) {
	var req dto.RedeemCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	result, err := h.redeemCodeService.Redeem(c.GetInt64("user_id"), req.Code)
	if err != nil { response.Error(c, err); return }
	response.SuccessWithMessage(c, "兑换成功", result)
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=dto.UserProfileResponse}
// @Router /api/v1/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	profile, err := h.userService.GetProfile(userID.(int64))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, profile)
}

// GetCreditLedgers returns the current user's own credit history.
func (h *UserHandler) GetCreditLedgers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}
	var req dto.CreditLedgerListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	items, total, err := h.userService.GetCreditLedgers(userID.(int64), req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, req.Page, req.PageSize)
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.UpdateUserRequest true "更新信息"
// @Success 200 {object} response.Response
// @Router /api/v1/users/profile [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	profile, err := h.userService.UpdateProfile(userID.(int64), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "更新成功", profile)
}

func (h *UserHandler) BindPhone(c *gin.Context) {
	var req dto.BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	profile, err := h.userService.BindPhone(c.GetInt64("user_id"), &req)
	if err != nil { response.Error(c, err); return }
	response.SuccessWithMessage(c, "手机号绑定成功", profile)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error())); return }
	if err := h.userService.ChangePassword(c.GetInt64("user_id"), c.GetString("session_id"), &req); err != nil { response.Error(c, err); return }
	response.SuccessWithMessage(c, "密码已修改，其他设备已退出登录", nil)
}

func (h *UserHandler) ListSessions(c *gin.Context) {
	sessions, err := h.userService.ListSessions(c.GetInt64("user_id"), c.GetString("session_id"))
	if err != nil { response.Error(c, err); return }; response.Success(c, sessions)
}

func (h *UserHandler) RevokeSession(c *gin.Context) {
	if err := h.userService.RevokeSession(c.GetInt64("user_id"), c.Param("id")); err != nil { response.Error(c, err); return }
	response.SuccessWithMessage(c, "设备已退出登录", nil)
}
