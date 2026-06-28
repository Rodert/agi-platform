package handler

import (
	"net/http"
	"strings"

	"agi-platform/server/internal/middleware"
	"agi-platform/server/internal/response"

	"github.com/gin-gonic/gin"
)

func currentUserID(c *gin.Context) (uint64, bool) {
	value, ok := c.Get(middleware.CurrentUserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint64)
	return userID, ok && userID > 0
}

func currentAPIKeyID(c *gin.Context) *uint64 {
	value, ok := c.Get(middleware.CurrentAPIKeyIDKey)
	if !ok {
		return nil
	}
	apiKeyID, ok := value.(uint64)
	if !ok || apiKeyID == 0 {
		return nil
	}
	return &apiKeyID
}

func currentAdminID(c *gin.Context) (uint64, bool) {
	value, ok := c.Get(middleware.CurrentAdminIDKey)
	if !ok {
		return 0, false
	}
	adminID, ok := value.(uint64)
	return adminID, ok && adminID > 0
}

func requireCurrentUserID(c *gin.Context) (uint64, bool) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "unauthorized")
		return 0, false
	}
	return userID, true
}

func requireCurrentAdminID(c *gin.Context) (uint64, bool) {
	adminID, ok := currentAdminID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "unauthorized")
		return 0, false
	}
	return adminID, true
}

func requestBaseURL(c *gin.Context) string {
	proto := firstHeaderValue(c.GetHeader("X-Forwarded-Proto"))
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := firstHeaderValue(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = c.Request.Host
	}
	if host == "" {
		return ""
	}
	return strings.ToLower(proto) + "://" + host
}

func firstHeaderValue(value string) string {
	if index := strings.Index(value, ","); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(value)
}
