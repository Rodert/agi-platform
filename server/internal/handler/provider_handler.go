package handler

import (
	"net/http"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type ProviderHandler interface {
	List(c *gin.Context)
}

type providerHandler struct {
	service service.ProviderService
}

func NewProviderHandler(service service.ProviderService) ProviderHandler {
	return &providerHandler{service: service}
}

func (h *providerHandler) List(c *gin.Context) {
	providers, err := h.service.ListEnabled(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, providers)
}
