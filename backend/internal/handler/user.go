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
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
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
