package flashcard

import (
	"mission-note/internal/pkg/openai"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterModule(rg *gin.RouterGroup, pool *pgxpool.Pool, aiClient *openai.Client) {
	repo := NewPgRepository(pool)
	svc := NewService(repo, aiClient)
	h := newHandler(svc)

	tags := rg.Group("/tags")
	{
		tags.POST("", h.createTag)
		tags.GET("", h.listTags)
		tags.POST("/:tag_id/flashcards/generate", h.generateFlashcards)
		tags.GET("/:tag_id/flashcards", h.listFlashcards)
	}

	cards := rg.Group("/flashcards")
	{
		cards.POST("", h.saveFlashcard)
		cards.GET("/:card_id", h.getFlashcard)
		cards.POST("/:card_id/sentences", h.addSentence)
		cards.POST("/:card_id/sentences/generate", h.generateSentences)
	}

	sentences := rg.Group("/sentences")
	{
		sentences.DELETE("/:sentence_id", h.deleteSentence)
	}
}
