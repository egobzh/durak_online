package middleware

import (
	"net/http"
	"strings"

	"durak-online/internal/domain"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	CtxUserID           = "userID"
)

func AuthMiddleware(authService domain.AuthService, userRepo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && strings.ToLower(parts[0]) == "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		userID, tokenVersion, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil || user.TokenVersion != tokenVersion {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session invalidated or user not found"})
			return
		}

		c.Set(CtxUserID, userID)
		c.Next()
	}
}
