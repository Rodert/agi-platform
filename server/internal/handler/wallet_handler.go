package handler

import (
	"net/http"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletHandler interface {
	List(c *gin.Context)
	AdminList(c *gin.Context)
}

type walletHandler struct {
	service service.WalletService
}

func NewWalletHandler(service service.WalletService) WalletHandler {
	return &walletHandler{service: service}
}

func (h *walletHandler) List(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	logs, err := h.service.ListForUser(c.Request.Context(), userID, queryInt(c, "limit", 50), queryInt(c, "offset", 0))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, logs)
}

func (h *walletHandler) AdminList(c *gin.Context) {
	logs, err := h.service.List(c.Request.Context(), queryInt(c, "limit", 100), queryInt(c, "offset", 0))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, logs)
}
