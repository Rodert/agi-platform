package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"agi-platform/server/internal/response"
	"agi-platform/server/internal/storage"

	"github.com/gin-gonic/gin"
)

type UploadHandler interface {
	Reference(c *gin.Context)
	Asset(c *gin.Context)
}

type uploadHandler struct {
	store storage.Store
}

type uploadReferenceResult struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

func NewUploadHandler(store storage.Store) UploadHandler {
	return &uploadHandler{store: store}
}

func (h *uploadHandler) Reference(c *gin.Context) {
	userID, ok := requireCurrentUserID(c)
	if !ok {
		return
	}

	kind := strings.TrimSpace(c.DefaultPostForm("kind", "image"))
	if !validUploadKind(kind) {
		response.Fail(c, http.StatusBadRequest, "invalid upload kind")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "file is required")
		return
	}
	if !validUploadFile(kind, file.Filename, file.Size) {
		response.Fail(c, http.StatusBadRequest, "unsupported file type or size")
		return
	}

	source, err := file.Open()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer source.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	key := storage.DatedKey("references", fmt.Sprintf("%d", userID), kind, fmt.Sprintf("%d%s", time.Now().UnixNano(), ext))
	mimeType := storage.MimeTypeFromFilename(file.Filename, file.Header.Get("Content-Type"))
	stored, err := h.store.Put(c.Request.Context(), key, source, file.Size, mimeType)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(c, uploadReferenceResult{
		URL:      stored.AppURL,
		Filename: file.Filename,
		MimeType: mimeType,
		Size:     stored.Size,
	})
}

func (h *uploadHandler) Asset(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	publicURL, ok := h.store.PublicURL(key)
	if !ok {
		response.Fail(c, http.StatusNotFound, "asset not found")
		return
	}
	c.Redirect(http.StatusFound, publicURL)
}

func validUploadKind(kind string) bool {
	return kind == "image" || kind == "video" || kind == "audio"
}

func validUploadFile(kind string, filename string, size int64) bool {
	if size <= 0 {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch kind {
	case "image":
		return size <= 20*1024*1024 && inStringSet(ext, ".png", ".jpg", ".jpeg", ".webp", ".gif")
	case "video":
		return size <= 300*1024*1024 && inStringSet(ext, ".mp4", ".mov", ".webm", ".m4v")
	case "audio":
		return size <= 80*1024*1024 && inStringSet(ext, ".mp3", ".wav", ".m4a", ".aac", ".ogg")
	default:
		return false
	}
}

func inStringSet(value string, items ...string) bool {
	for _, item := range items {
		if value == item {
			return true
		}
	}
	return false
}
