package middleware

import (
	"net/http"
	"strings"

	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/response"
	"agi-platform/server/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	CurrentUserIDKey   = "current_user_id"
	CurrentAdminIDKey  = "current_admin_id"
	CurrentAPIKeyIDKey = "current_api_key_id"
	AuthTypeKey        = "auth_type"
)

func UserAuth(authManager auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			response.Fail(c, http.StatusUnauthorized, "missing token")
			c.Abort()
			return
		}

		claims, err := authManager.ParseUserToken(token)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set(CurrentUserIDKey, claims.UserID)
		c.Set(AuthTypeKey, "jwt")
		c.Next()
	}
}

func AdminAuth(authManager auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			response.Fail(c, http.StatusUnauthorized, "missing token")
			c.Abort()
			return
		}

		claims, err := authManager.ParseAdminToken(token)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set(CurrentAdminIDKey, claims.UserID)
		c.Set(AuthTypeKey, "admin_jwt")
		c.Next()
	}
}

func APIKeyAuth(apiKeyService service.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		plain := bearerToken(c.GetHeader("Authorization"))
		if plain == "" {
			plain = c.GetHeader("X-API-Key")
		}
		if plain == "" {
			response.Fail(c, http.StatusUnauthorized, "missing api key")
			c.Abort()
			return
		}

		key, err := apiKeyService.Authenticate(c.Request.Context(), plain)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "invalid api key")
			c.Abort()
			return
		}

		c.Set(CurrentUserIDKey, key.UserID)
		c.Set(CurrentAPIKeyIDKey, key.ID)
		c.Set(AuthTypeKey, "api_key")
		c.Next()
	}
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
