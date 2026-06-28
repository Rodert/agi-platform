package handler

import (
	"net/http"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

type VideoHandler interface {
	ListModels(c *gin.Context)
	Submit(c *gin.Context)
	List(c *gin.Context)
	Get(c *gin.Context)
	AdminList(c *gin.Context)
	AdminGet(c *gin.Context)
}

type videoHandler struct {
	service service.VideoService
}

type submitVideoRequest struct {
	Model       string   `json:"model" binding:"required"`
	Prompt      string   `json:"prompt" binding:"required"`
	Seconds     int      `json:"seconds"`
	AspectRatio string   `json:"aspect_ratio"`
	Images      []string `json:"images"`
	Videos      []string `json:"videos"`
	Audios      []string `json:"audios"`
}

func NewVideoHandler(service service.VideoService) VideoHandler {
	return &videoHandler{service: service}
}

func (h *videoHandler) ListModels(c *gin.Context) {
	models, err := h.service.ListModels(c.Request.Context(), queryInt(c, "limit", 50), queryInt(c, "offset", 0))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, models)
}

func (h *videoHandler) Submit(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}
	var req submitVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.service.Submit(c.Request.Context(), service.SubmitVideoRequest{
		UserID:      userID,
		Source:      "web",
		Model:       req.Model,
		Prompt:      req.Prompt,
		Seconds:     req.Seconds,
		AspectRatio: req.AspectRatio,
		Images:      req.Images,
		Videos:      req.Videos,
		Audios:      req.Audios,
		AppBaseURL:  requestBaseURL(c),
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Created(c, result)
}

func (h *videoHandler) List(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}
	tasks, err := h.service.ListForUser(c.Request.Context(), userID, queryInt(c, "limit", 50), queryInt(c, "offset", 0))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, tasks)
}

func (h *videoHandler) Get(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}
	result, err := h.service.GetForUser(c.Request.Context(), userID, c.Param("task_no"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *videoHandler) AdminList(c *gin.Context) {
	tasks, err := h.service.ListFiltered(c.Request.Context(), taskFilterFromQuery(c), queryInt(c, "limit", 50), queryInt(c, "offset", 0))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, tasks)
}

func (h *videoHandler) AdminGet(c *gin.Context) {
	result, err := h.service.Get(c.Request.Context(), c.Param("task_no"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, result)
}
