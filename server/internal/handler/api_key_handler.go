package handler

import (
	"net/http"
	"strconv"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type APIKeyHandler interface {
	List(c *gin.Context)
	Create(c *gin.Context)
	Revoke(c *gin.Context)
}

type apiKeyHandler struct {
	service service.APIKeyService
}

type createAPIKeyRequest struct {
	Name string `json:"name"`
}

func NewAPIKeyHandler(service service.APIKeyService) APIKeyHandler {
	return &apiKeyHandler{service: service}
}

func (h *apiKeyHandler) List(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	keys, err := h.service.List(c.Request.Context(), userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, keys)
}

func (h *apiKeyHandler) Create(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.Create(c.Request.Context(), userID, req.Name)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Created(c, result)
}

func (h *apiKeyHandler) Revoke(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid api key id")
		return
	}

	if err := h.service.Revoke(c.Request.Context(), userID, id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}
