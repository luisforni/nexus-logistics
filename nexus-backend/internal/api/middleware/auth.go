package middleware

import (
"net/http"
"strings"

"github.com/gin-gonic/gin"
"github.com/luisforni/nexus-logistics/backend/internal/domain"
)

const (
ContextUserID = "user_id"
ContextRole   = "user_role"
ContextEmail  = "user_email"
)

type tokenValidator interface {
ValidateToken(token string) (*domain.Claims, error)
}

func JWTAuth(authSvc tokenValidator) gin.HandlerFunc {
return func(c *gin.Context) {
header := c.GetHeader("Authorization")
if header == "" {
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
return
}

parts := strings.SplitN(header, " ", 2)
if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "malformed authorization header"})
return
}

claims, err := authSvc.ValidateToken(parts[1])
if err != nil {
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
return
}

c.Set(ContextUserID, claims.UserID)
c.Set(ContextRole, string(claims.Role))
c.Set(ContextEmail, claims.Email)
c.Next()
}
}

func RequireRole(roles ...string) gin.HandlerFunc {
allowed := make(map[string]struct{}, len(roles))
for _, r := range roles {
allowed[r] = struct{}{}
}

return func(c *gin.Context) {
role, _ := c.Get(ContextRole)
if _, ok := allowed[role.(string)]; !ok {
c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
return
}
c.Next()
}
}
