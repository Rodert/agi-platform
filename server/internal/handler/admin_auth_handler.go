package handler

import (
	"net/http"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminAuthHandler interface {
	Login(c *gin.Context)
	Me(c *gin.Context)
	ListUsers(c *gin.Context)
	AdjustUserCredits(c *gin.Context)
}

type adminAuthHandler struct {
	service service.AdminService
}

type adminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type adjustUserCreditsRequest struct {
	Amount int64  `json:"amount" binding:"required"`
	Remark string `json:"remark"`
}

func NewAdminAuthHandler(service service.AdminService) AdminAuthHandler {
	return &adminAuthHandler{service: service}
}

func (h *adminAuthHandler) Login(c *gin.Context) {
	var req adminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.Login(c.Request.Context(), service.AdminLoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *adminAuthHandler) Me(c *gin.Context) {
	adminID, ok := requireCurrentAdminID(c)
	if !ok {
		return
	}

	admin, err := h.service.Me(c.Request.Context(), adminID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, admin)
}

func (h *adminAuthHandler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers(c.Request.Context(), queryInt(c, "limit", 20), queryInt(c, "offset", 0))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, users)
}

func (h *adminAuthHandler) AdjustUserCredits(c *gin.Context) {
	adminID, ok := requireCurrentAdminID(c)
	if !ok {
		return
	}
	userID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req adjustUserCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.AdjustUserCredits(c.Request.Context(), service.AdjustUserCreditsRequest{
		AdminID: adminID,
		UserID:  userID,
		Amount:  req.Amount,
		Remark:  req.Remark,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, user)
}
