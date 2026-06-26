package handler

import (
	"net/http"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type ImageModelHandler interface {
	List(c *gin.Context)
}

type imageModelHandler struct {
	service service.ImageModelService
}

func NewImageModelHandler(service service.ImageModelService) ImageModelHandler {
	return &imageModelHandler{service: service}
}

func (h *imageModelHandler) List(c *gin.Context) {
	models, err := h.service.ListEnabled(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, models)
}
