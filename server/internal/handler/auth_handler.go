package handler

import (
	"errors"
	"net/http"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Me(c *gin.Context)
}

type authHandler struct {
	service service.AuthService
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(service service.AuthService) AuthHandler {
	return &authHandler{service: service}
}

func (h *authHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.Register(c.Request.Context(), service.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Created(c, result)
}

func (h *authHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.Login(c.Request.Context(), service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Fail(c, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeServiceError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *authHandler) Me(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	user, err := h.service.Me(c.Request.Context(), userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, user)
}
