// Package middleware holds cross-cutting Gin middleware shared across feature modules.
package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"mission-note/internal/core/response"
)

const userIDContextKey = "auth.userID"

// RequireAuth validates the Bearer JWT on incoming requests and stores the
// authenticated user's ID in the request context for downstream handlers.
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || tokenStr == "" {
			response.Error(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid token subject")
			c.Abort()
			return
		}

		c.Set(userIDContextKey, userID)
		c.Next()
	}
}

// UserID returns the authenticated user's ID stored by RequireAuth.
func UserID(c *gin.Context) (int64, bool) {
	v, exists := c.Get(userIDContextKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
