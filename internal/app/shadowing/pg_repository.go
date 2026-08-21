package shadowing

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository creates a PostgreSQL-backed Repository.
func NewPgRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) FindSentenceForUser(ctx context.Context, sentenceID, userID int64) (SentenceRef, error) {
	const query = `
		SELECT fs.id, fs.sentence_text
		FROM flashcard_sentences fs
		JOIN flashcards f ON f.id = fs.flashcard_id
		WHERE fs.id = $1 AND f.user_id = $2`

	var s SentenceRef
	err := r.pool.QueryRow(ctx, query, sentenceID, userID).Scan(&s.ID, &s.Text)
	if err != nil {
		return SentenceRef{}, err
	}
	return s, nil
}

func (r *pgRepository) SaveAttempt(ctx context.Context, a Attempt) (Attempt, error) {
	const query = `
		INSERT INTO shadowing_attempts (user_id, flashcard_sentence_id, transcript, score, correct_words, mispronounced_words)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query, a.UserID, a.SentenceID, a.Transcript, a.Score, a.CorrectWords, a.MispronouncedWords).
		Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return Attempt{}, err
	}
	return a, nil
}
