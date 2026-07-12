package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/pkg/config"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/jwt"
	"github.com/javapub/agi-platform-backend/pkg/response"
)

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// 解析管理员 token
		claims, err := jwt.ParseAdminToken(parts[1], cfg)
		if err != nil {
			response.Error(c, errors.NewWithDetails(errors.ErrCodeUnauthorized, "Token无效", err.Error()))
			c.Abort()
			return
		}

		// 将管理员信息存入上下文
		c.Set("admin_id", claims.AdminID)
		c.Set("admin_username", claims.Username)
		c.Set("admin_role", claims.Role)

		c.Next()
	}
}

// RequireRole 要求特定角色
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("admin_role")
		if !exists {
			response.Error(c, errors.ErrForbidden)
			c.Abort()
			return
		}

		roleStr := role.(string)

		// super_admin 有所有权限
		if roleStr == "super_admin" {
			c.Next()
			return
		}

		// 检查是否在允许的角色列表中
		for _, r := range roles {
			if r == roleStr {
				c.Next()
				return
			}
		}

		response.Error(c, errors.ErrForbidden)
		c.Abort()
	}
}
