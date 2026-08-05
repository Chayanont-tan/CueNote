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

func (r *pgRepository) FindSentenceByID(ctx context.Context, id int64) (Sentence, error) {
	const query = `
		SELECT id, text, audio_url, timestamps, created_at
		FROM shadowing_sentences
		WHERE id = $1`

	var s Sentence
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&s.ID, &s.Text, &s.AudioURL, &s.Timestamps, &s.CreatedAt)
	if err != nil {
		return Sentence{}, err
	}
	return s, nil
}

func (r *pgRepository) SaveAttempt(ctx context.Context, a Attempt) (Attempt, error) {
	const query = `
		INSERT INTO shadowing_attempts (user_id, sentence_id, audio_url, score)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query, a.UserID, a.SentenceID, a.AudioURL, a.Score).
		Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return Attempt{}, err
	}
	return a, nil
}
