package middleware

import (
	"net/http"
	"strings"

	"productcore/internal/config"
	"productcore/internal/pkg/authcontext"
	jwtmgr "productcore/internal/pkg/jwt"
	"productcore/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func AdminAuth(cfg *config.AuthConfig, jwt *jwtmgr.Manager) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Set(authcontext.ContextTenant, uint64(1))
			c.Next()
		}
	}
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "请先登录")
			c.Abort()
			return
		}
		if jwt == nil {
			response.Fail(c, http.StatusUnauthorized, "JWT 未配置")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		claims, err := jwt.ParseAccess(token)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		if claims.TenantID == 0 {
			response.Fail(c, http.StatusUnauthorized, "请选择租户")
			c.Abort()
			return
		}
		c.Set(authcontext.ContextClaims, claims)
		c.Set(authcontext.ContextTenant, claims.TenantID)
		c.Set(authcontext.ContextUser, claims.UserID)
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
