package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/pkg/logger"
	"github.com/javapub/agi-platform-backend/pkg/response"
	"go.uber.org/zap"
)

// RecoveryMiddleware 错误恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				c.JSON(http.StatusInternalServerError, response.Response{
					Success: false,
					Error: &response.ErrorInfo{
						Code:    1005,
						Message: "服务器内部错误",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
