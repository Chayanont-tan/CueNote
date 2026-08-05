// Package auth implements identity & account management.
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterModule wires the repository, service and handler together and
// registers the auth feature's routes on the given router group.
func RegisterModule(rg *gin.RouterGroup, pool *pgxpool.Pool, jwtSecret string) {
	repo := NewPgRepository(pool)
	svc := NewService(repo, jwtSecret)
	h := newHandler(svc)

	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.register)
		auth.POST("/login", h.login)
	}
}
