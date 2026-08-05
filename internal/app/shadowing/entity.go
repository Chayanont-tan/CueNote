package shadowing

import "time"

// Sentence is a reference sentence with AI-generated audio and word timestamps
// used to drive the karaoke-style highlight during playback.
type Sentence struct {
	ID         int64     `db:"id"`
	Text       string    `db:"text"`
	AudioURL   string    `db:"audio_url"`
	Timestamps []byte    `db:"timestamps"` // JSON-encoded [{word, start_ms, end_ms}, ...]
	CreatedAt  time.Time `db:"created_at"`
}

// Attempt is a user's recorded shadowing attempt and its pronunciation score.
type Attempt struct {
	ID         int64     `db:"id"`
	UserID     int64     `db:"user_id"`
	SentenceID int64     `db:"sentence_id"`
	AudioURL   string    `db:"audio_url"`
	Score      int       `db:"score"`
	CreatedAt  time.Time `db:"created_at"`
}
