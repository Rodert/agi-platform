package handler

import (
	"errors"
	"net/http"
	"strconv"

	"agi-platform/server/internal/provider"
	"agi-platform/server/internal/repository"
	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type ImageTaskHandler interface {
	Generate(c *gin.Context)
	List(c *gin.Context)
	Get(c *gin.Context)
	AdminList(c *gin.Context)
	AdminGet(c *gin.Context)
}

type imageTaskHandler struct {
	service service.ImageTaskService
}

type generateImageRequest struct {
	Model           string                  `json:"model" binding:"required"`
	Prompt          string                  `json:"prompt" binding:"required"`
	NegativePrompt  string                  `json:"negative_prompt"`
	Size            string                  `json:"size" binding:"required"`
	N               int                     `json:"n"`
	ReferenceImages []referenceImageRequest `json:"reference_images"`
}

type referenceImageRequest struct {
	URL      string `json:"url"`
	DataURL  string `json:"data_url"`
	Filename string `json:"filename"`
}

func NewImageTaskHandler(service service.ImageTaskService) ImageTaskHandler {
	return &imageTaskHandler{service: service}
}

func (h *imageTaskHandler) Generate(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	var req generateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.Submit(c.Request.Context(), service.GenerateImageRequest{
		UserID:          userID,
		Source:          "web",
		Model:           req.Model,
		Prompt:          req.Prompt,
		NegativePrompt:  req.NegativePrompt,
		Size:            req.Size,
		N:               req.N,
		ReferenceImages: toProviderReferenceImages(req.ReferenceImages),
		AppBaseURL:      requestBaseURL(c),
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}

	response.Created(c, result)
}

func toProviderReferenceImages(items []referenceImageRequest) []provider.ReferenceImage {
	images := make([]provider.ReferenceImage, 0, len(items))
	for _, item := range items {
		images = append(images, provider.ReferenceImage{
			URL:      item.URL,
			DataURL:  item.DataURL,
			Filename: item.Filename,
		})
	}
	return images
}

func (h *imageTaskHandler) Get(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	taskNo := c.Param("task_no")
	result, err := h.service.GetForUser(c.Request.Context(), userID, taskNo)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *imageTaskHandler) List(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	tasks, err := h.service.ListForUser(c.Request.Context(), userID, queryInt(c, "limit", 20), queryInt(c, "offset", 0))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, tasks)
}

func (h *imageTaskHandler) AdminList(c *gin.Context) {
	limit := queryInt(c, "limit", 20)
	offset := queryInt(c, "offset", 0)
	filter := taskFilterFromQuery(c)

	tasks, err := h.service.ListFiltered(c.Request.Context(), filter, limit, offset)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, tasks)
}

func (h *imageTaskHandler) AdminGet(c *gin.Context) {
	taskNo := c.Param("task_no")
	result, err := h.service.Get(c.Request.Context(), taskNo)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, result)
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidRequest):
		response.Fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrEmailAlreadyExists):
		response.Fail(c, http.StatusConflict, "email already exists")
	case errors.Is(err, service.ErrInvalidCredentials):
		response.Fail(c, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, service.ErrInsufficientCredits):
		response.Fail(c, http.StatusPaymentRequired, "insufficient credits")
	case errors.Is(err, repository.ErrNotFound):
		response.Fail(c, http.StatusNotFound, "record not found")
	default:
		response.Fail(c, http.StatusInternalServerError, err.Error())
	}
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func taskFilterFromQuery(c *gin.Context) repository.TaskFilter {
	var filter repository.TaskFilter
	if id := queryInt(c, "id", 0); id > 0 {
		filter.ID = uint64(id)
	}
	filter.TaskNo = c.Query("task_no")
	filter.Keyword = c.Query("keyword")
	filter.Status = c.Query("status")
	return filter
}

func parseUintParam(c *gin.Context, key string) (uint64, bool) {
	value := c.Param(key)
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid "+key)
		return 0, false
	}
	return parsed, true
}
