package handler

import (
	"net/http"
	"time"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type OpenAIImageHandler interface {
	Generate(c *gin.Context)
}

type openAIImageHandler struct {
	service service.ImageTaskService
}

type openAIImageRequest struct {
	Model  string `json:"model" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
	Size   string `json:"size" binding:"required"`
	N      int    `json:"n"`
}

type openAIImageResponse struct {
	Created int64             `json:"created"`
	Data    []openAIImageData `json:"data"`
	Usage   openAIImageUsage  `json:"usage"`
	TaskID  string            `json:"task_id"`
}

type openAIImageData struct {
	URL string `json:"url"`
}

type openAIImageUsage struct {
	Credits int64 `json:"credits"`
}

func NewOpenAIImageHandler(service service.ImageTaskService) OpenAIImageHandler {
	return &openAIImageHandler{service: service}
}

func (h *openAIImageHandler) Generate(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	var req openAIImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.Generate(c.Request.Context(), service.GenerateImageRequest{
		UserID:   userID,
		APIKeyID: currentAPIKeyID(c),
		Source:   "api",
		Model:    req.Model,
		Prompt:   req.Prompt,
		Size:     req.Size,
		N:        req.N,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}

	data := make([]openAIImageData, 0, len(result.Images))
	for _, image := range result.Images {
		data = append(data, openAIImageData{URL: image.URL})
	}

	response.OK(c, openAIImageResponse{
		Created: time.Now().Unix(),
		Data:    data,
		Usage: openAIImageUsage{
			Credits: result.Task.CreditsUsed,
		},
		TaskID: result.Task.TaskNo,
	})
}
