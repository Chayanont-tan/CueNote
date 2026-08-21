package shadowing

import "time"

// SentenceRef is a lightweight reference to a flashcard sentence the user
// can shadow-read — shadowing reuses the sentences a user already wrote or
// picked while building their flashcards instead of a separate content set.
type SentenceRef struct {
	ID   int64  `db:"id"`
	Text string `db:"sentence_text"`
}

// Attempt is a user's recorded shadowing attempt and its pronunciation score.
type Attempt struct {
	ID                 int64     `db:"id"`
	UserID             int64     `db:"user_id"`
	SentenceID         int64     `db:"flashcard_sentence_id"`
	Transcript         string    `db:"transcript"`
	Score              int       `db:"score"`
	CorrectWords       []byte    `db:"correct_words"`       // JSON-encoded []string
	MispronouncedWords []byte    `db:"mispronounced_words"` // JSON-encoded []string
	CreatedAt          time.Time `db:"created_at"`
}
