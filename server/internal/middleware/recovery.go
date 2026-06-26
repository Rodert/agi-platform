package middleware

import (
	"net/http"

	"agi-platform/server/internal/response"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, _ interface{}) {
		response.Fail(c, http.StatusInternalServerError, "internal server error")
	})
}
