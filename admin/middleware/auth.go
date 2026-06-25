package middleware

import (
	"net/http"
	"strings"

	"productcore/internal/config"
	"productcore/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func AdminAuth(cfg *config.AuthConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}
	expected := "Bearer " + cfg.AdminToken
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth != expected {
			response.Fail(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

// OptionalAuth allows open routes but validates token when present.
func OptionalBearer(cfg *config.AuthConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth != "" && !strings.HasPrefix(auth, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "invalid authorization header")
			c.Abort()
			return
		}
		c.Next()
	}
}
