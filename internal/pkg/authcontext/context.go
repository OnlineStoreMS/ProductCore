package authcontext

import (
	"strings"

	jwtmgr "productcore/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const (
	ContextClaims = "auth_claims"
	ContextTenant = "tenant_id"
	ContextUser   = "user_id"
)

func TenantID(c *gin.Context) uint64 {
	if v, ok := c.Get(ContextTenant); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	if claims := Claims(c); claims != nil {
		return claims.TenantID
	}
	return 0
}

func UserID(c *gin.Context) uint64 {
	if claims := Claims(c); claims != nil {
		return claims.UserID
	}
	return 0
}

func Claims(c *gin.Context) *jwtmgr.Claims {
	v, ok := c.Get(ContextClaims)
	if !ok {
		return nil
	}
	claims, ok := v.(*jwtmgr.Claims)
	if !ok {
		return nil
	}
	return claims
}

func HasPerm(c *gin.Context, code string) bool {
	claims := Claims(c)
	if claims == nil {
		return false
	}
	if claims.IsPlatform {
		return true
	}
	for _, p := range claims.Permissions {
		if p == code || p == "*" {
			return true
		}
	}
	return false
}

func RequirePerm(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasPerm(c, code) {
			c.AbortWithStatusJSON(403, gin.H{"code": 403, "message": "权限不足"})
			return
		}
		c.Next()
	}
}

func BearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
