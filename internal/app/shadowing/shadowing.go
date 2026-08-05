// Package shadowing implements Feature 3: shadowing player + pronunciation scoring.
package shadowing

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"mission-note/internal/infra/storage"
	"mission-note/internal/pkg/openai"
)

// RegisterModule wires the repository, service and handler together and
// registers the shadowing feature's routes on the given router group.
func RegisterModule(rg *gin.RouterGroup, pool *pgxpool.Pool, storageClient storage.Storage, aiClient *openai.Client) {
	repo := NewPgRepository(pool)
	svc := NewService(repo, storageClient, aiClient)
	h := newHandler(svc)

	sentences := rg.Group("/shadowing")
	{
		sentences.GET("/sentences/:id", h.getSentence)
		sentences.POST("/attempts", h.submitAttempt)
	}
}
