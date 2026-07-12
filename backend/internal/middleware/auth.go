package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/pkg/config"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/jwt"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// 检查格式: Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// 解析 token
		claims, err := jwt.ParseToken(parts[1], cfg)
		if err != nil {
			response.Error(c, errors.NewWithDetails(errors.ErrCodeUnauthorized, "Token无效", err.Error()))
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// OptionalAuthMiddleware 可选的认证中间件（不强制登录）
func OptionalAuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			claims, err := jwt.ParseToken(parts[1], cfg)
			if err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("email", claims.Email)
			}
		}

		c.Next()
	}
}
