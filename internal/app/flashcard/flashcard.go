// Package flashcard implements the flashcard feature: tags (topics/missions),
// the vocabulary catalog behind them, the flashcards generated from those
// words (AI image + AI sentences), and the sentences a user adds to a card.
package flashcard

import (
	"mission-note/internal/pkg/openai"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterModule wires the repository, service and handler together and
// registers the flashcard feature's routes on the given router group.
func RegisterModule(rg *gin.RouterGroup, pool *pgxpool.Pool, aiClient *openai.Client) {
	repo := NewPgRepository(pool)
	svc := NewService(repo, aiClient)
	h := newHandler(svc)

	tags := rg.Group("/tags")
	{
		tags.POST("", h.createTag)                                      // สร้าง tag ใหม่ (ยังไม่เจนคำศัพท์)
		tags.GET("", h.listTags)                                        // หน้าแรก: ?limit=10 / หน้า All Tags: ไม่ใส่ limit
		tags.GET("/:tag_id", h.getTag)                                  // รายละเอียด tag เดียว
		tags.POST("/:tag_id/flashcards/generate", h.generateFlashcards) // สั่ง AI เจน flashcard ใหม่ในนี้
		tags.GET("/:tag_id/flashcards", h.listFlashcards)               // รายการ flashcard ใน tag (?level=A1 กรองได้)
	}

	cards := rg.Group("/flashcards")
	{
		cards.GET("/:card_id", h.getFlashcard)                          // รายละเอียด flashcard ใบเดียว
		cards.POST("/:card_id/sentences", h.addSentence)                // เพิ่มประโยคที่พิมพ์เอง/เลือกจาก AI
		cards.POST("/:card_id/sentences/generate", h.generateSentences) // ให้ AI เจนประโยคใหม่ 3 ประโยค
	}

	sentences := rg.Group("/sentences")
	{
		sentences.DELETE("/:sentence_id", h.deleteSentence) // ลบประโยคที่เคยเพิ่มไว้
	}
}
