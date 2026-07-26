package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=dto.AuthResponse}
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	result, err := h.authService.Register(&req, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// Login 用户登录
// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=dto.AuthResponse}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	result, err := h.authService.Login(&req, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// SendCode 发送验证码
// @Summary 发送验证码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.SendCodeRequest true "验证码请求"
// @Success 200 {object} response.Response
// @Router /api/v1/auth/send-code [post]
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req dto.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "参数错误", err.Error()))
		return
	}

	if err := h.authService.SendCode(&req); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "验证码已发送，请查收邮箱", nil)
}
