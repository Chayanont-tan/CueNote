package flashcard

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mission-note/internal/pkg/openai"
)

var ErrNotFound = errors.New("not found")

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewPgRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) CreateTag(ctx context.Context, userID int64, name string) (Tag, error) {
	const query = `
		INSERT INTO tags (name, created_by)
		VALUES ($1, $2)
		ON CONFLICT (created_by, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, created_at;
	`
	var t Tag
	err := r.pool.QueryRow(ctx, query, name, userID).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err != nil {
		return Tag{}, fmt.Errorf("failed to create tag: %w", err)
	}
	return t, nil
}

func (r *pgRepository) ListTagsForUser(ctx context.Context, userID int64, limit int) ([]TagResponse, error) {
	const query = `
		SELECT
			t.id,
			t.name,
			COUNT(DISTINCT vt.vocabulary_id) AS total_words
		FROM tags t
		LEFT JOIN vocabulary_tags vt ON vt.tag_id = t.id
		WHERE t.created_by = $1
		GROUP BY t.id, t.name, t.created_at
		ORDER BY t.created_at DESC
		LIMIT $2;
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags for user: %w", err)
	}
	defer rows.Close()

	result := []TagResponse{}
	for rows.Next() {
		var t TagResponse
		if err := rows.Scan(&t.ID, &t.Name, &t.TotalWords); err != nil {
			return nil, fmt.Errorf("failed to scan tag row: %w", err)
		}
		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return result, nil
}

func (r *pgRepository) GetTagByID(ctx context.Context, tagID int64, userID int64) (TagResponse, error) {
	const query = `
		SELECT
			t.id,
			t.name,
			COUNT(DISTINCT vt.vocabulary_id) AS total_words
		FROM tags t
		LEFT JOIN vocabulary_tags vt ON vt.tag_id = t.id
		WHERE t.id = $1 AND t.created_by = $2
		GROUP BY t.id, t.name;
	`

	var t TagResponse
	err := r.pool.QueryRow(ctx, query, tagID, userID).Scan(&t.ID, &t.Name, &t.TotalWords)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TagResponse{}, ErrNotFound
		}
		return TagResponse{}, fmt.Errorf("failed to get tag: %w", err)
	}
	return t, nil
}

func (r *pgRepository) FindVocabByWord(ctx context.Context, word string) (Vocabulary, bool, error) {
	const query = `
		SELECT id, word, part_of_speech, meaning_th, level, created_at
		FROM vocabularies
		WHERE LOWER(word) = LOWER($1)
		LIMIT 1;
	`

	var v Vocabulary
	err := r.pool.QueryRow(ctx, query, word).Scan(&v.ID, &v.Word, &v.PartOfSpeech, &v.MeaningTH, &v.Level, &v.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Vocabulary{}, false, nil
		}
		return Vocabulary{}, false, fmt.Errorf("failed to find vocab by word: %w", err)
	}
	return v, true, nil
}

func (r *pgRepository) ListVocabWordsForTag(ctx context.Context, tagID int64) ([]string, error) {
	const query = `
		SELECT v.word
		FROM vocabularies v
		JOIN vocabulary_tags vt ON v.id = vt.vocabulary_id
		WHERE vt.tag_id = $1;
	`

	rows, err := r.pool.Query(ctx, query, tagID)
	if err != nil {
		return nil, fmt.Errorf("failed to query vocab words for tag: %w", err)
	}
	defer rows.Close()

	var words []string
	for rows.Next() {
		var w string
		if err := rows.Scan(&w); err != nil {
			return nil, fmt.Errorf("failed to scan vocab word: %w", err)
		}
		words = append(words, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return words, nil
}

func (r *pgRepository) SaveVocabulariesForTag(ctx context.Context, tagID int64, items []openai.OpenAIVocabItem) ([]Vocabulary, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var saved []Vocabulary
	for _, item := range items {
		level := item.Level
		if level == "" {
			level = "A1"
		}

		var v Vocabulary
		const vocabQuery = `
			INSERT INTO vocabularies (word, part_of_speech, meaning_th, level)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (word, part_of_speech) DO UPDATE
				SET meaning_th = EXCLUDED.meaning_th
			RETURNING id, word, part_of_speech, meaning_th, level;
		`
		err := tx.QueryRow(ctx, vocabQuery, item.Word, item.PartOfSpeech, item.MeaningTH, level).Scan(
			&v.ID, &v.Word, &v.PartOfSpeech, &v.MeaningTH, &v.Level,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert vocabulary (%s): %w", item.Word, err)
		}

		const relationQuery = `
			INSERT INTO vocabulary_tags (vocabulary_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING;
		`
		if _, err := tx.Exec(ctx, relationQuery, v.ID, tagID); err != nil {
			return nil, fmt.Errorf("failed to link vocab to tag: %w", err)
		}

		saved = append(saved, v)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return saved, nil
}

func (r *pgRepository) CreateFlashcardWithSentences(ctx context.Context, userID int64, vocabularyID int64, imageURL string, aiSentences []string) (Flashcard, []AISuggestedSentence, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Flashcard{}, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var flashcard Flashcard
	const flashcardQuery = `
		INSERT INTO flashcards (user_id, vocabulary_id, image_url)
		VALUES ($1, $2, $3)
		RETURNING id::text, user_id, vocabulary_id, image_url, created_at, updated_at;
	`
	err = tx.QueryRow(ctx, flashcardQuery, userID, vocabularyID, imageURL).Scan(
		&flashcard.ID, &flashcard.UserID, &flashcard.VocabularyID, &flashcard.ImageURL,
		&flashcard.CreatedAt, &flashcard.UpdatedAt,
	)
	if err != nil {
		return Flashcard{}, nil, fmt.Errorf("failed to insert flashcard: %w", err)
	}

	savedSentences := make([]AISuggestedSentence, 0, len(aiSentences))
	for _, sentence := range aiSentences {
		saved := AISuggestedSentence{FlashcardID: flashcard.ID, SentenceText: sentence}
		const sentenceQuery = `
			INSERT INTO ai_suggested_sentences (flashcard_id, sentence_text)
			VALUES ($1, $2)
			RETURNING id, created_at;
		`
		err = tx.QueryRow(ctx, sentenceQuery, flashcard.ID, sentence).Scan(&saved.ID, &saved.CreatedAt)
		if err != nil {
			return Flashcard{}, nil, fmt.Errorf("failed to insert ai suggested sentence: %w", err)
		}
		savedSentences = append(savedSentences, saved)
	}

	if err := tx.Commit(ctx); err != nil {
		return Flashcard{}, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return flashcard, savedSentences, nil
}

func (r *pgRepository) ListFlashcardsForTag(ctx context.Context, tagID int64, userID int64) ([]FlashcardForTag, error) {
	// เตรียม คำสั้ง sql
	query := `
		SELECT f.id, v.word, v.part_of_speech, v.meaning_th, v.level
		FROM flashcards f
		JOIN vocabularies v ON v.id = f.vocabulary_id
		JOIN vocabulary_tags vt ON vt.vocabulary_id = v.id
		WHERE vt.tag_id = $1 AND f.user_id = $2
		ORDER BY f.created_at DESC;
	`
	// ส่งคำสั้ง sql ไปที่ database
	rows, err := r.pool.Query(ctx, query, tagID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query flashcards for tag: %w", err)
	}
	defer rows.Close()
	// สร้าง slice มาเก็บ

	result := []FlashcardForTag{}

	for rows.Next() {
		var f FlashcardForTag
		if err := rows.Scan(&f.ID, &f.Word, &f.PartOfSpeech, &f.MeaningTH, &f.Level); err != nil {
			return nil, fmt.Errorf("failed to scan flashcard row: %w", err)
		}
		result = append(result, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}
	return result, nil
}

func (r *pgRepository) GetFlashcardByID(ctx context.Context, flashcardID string, userID int64) (FlashcardResponse, error) {
	const query = `
		SELECT f.id, v.word, v.part_of_speech, v.meaning_th, v.level, f.image_url, f.created_at
		FROM flashcards f
		JOIN vocabularies v ON v.id = f.vocabulary_id
		WHERE f.id = $1 AND f.user_id = $2;
	`

	var f FlashcardResponse
	err := r.pool.QueryRow(ctx, query, flashcardID, userID).Scan(
		&f.ID, &f.Word, &f.PartOfSpeech, &f.MeaningTH, &f.Level, &f.ImageURL, &f.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FlashcardResponse{}, ErrNotFound
		}
		return FlashcardResponse{}, fmt.Errorf("failed to get flashcard: %w", err)
	}

	result := []FlashcardResponse{f}
	if err := r.attachSentences(ctx, result); err != nil {
		return FlashcardResponse{}, err
	}

	return result[0], nil
}

func (r *pgRepository) attachSentences(ctx context.Context, cards []FlashcardResponse) error {
	for i := range cards {
		aiRows, err := r.pool.Query(ctx, `
			SELECT sentence_text FROM ai_suggested_sentences
			WHERE flashcard_id = $1 ORDER BY created_at;
		`, cards[i].ID)
		if err != nil {
			return fmt.Errorf("failed to query ai suggested sentences: %w", err)
		}
		aiSentences := []string{}
		for aiRows.Next() {
			var s string
			if err := aiRows.Scan(&s); err != nil {
				aiRows.Close()
				return fmt.Errorf("failed to scan ai suggested sentence: %w", err)
			}
			aiSentences = append(aiSentences, s)
		}
		aiRows.Close()
		cards[i].AISuggestedSentences = aiSentences

		sentRows, err := r.pool.Query(ctx, `
			SELECT id, sentence_text, source, created_at FROM flashcard_sentences
			WHERE flashcard_id = $1 ORDER BY created_at;
		`, cards[i].ID)
		if err != nil {
			return fmt.Errorf("failed to query flashcard sentences: %w", err)
		}
		sentences := []SentenceResponse{}
		for sentRows.Next() {
			var s SentenceResponse
			if err := sentRows.Scan(&s.ID, &s.SentenceText, &s.Source, &s.CreatedAt); err != nil {
				sentRows.Close()
				return fmt.Errorf("failed to scan flashcard sentence: %w", err)
			}
			sentences = append(sentences, s)
		}
		sentRows.Close()
		cards[i].Sentences = sentences
	}
	return nil
}

func (r *pgRepository) AddSentence(ctx context.Context, flashcardID string, userID int64, text string, source string) (SentenceResponse, error) {
	const query = `
		INSERT INTO flashcard_sentences (flashcard_id, sentence_text, source)
		SELECT f.id, $2, $3
		FROM flashcards f
		WHERE f.id = $1 AND f.user_id = $4
		RETURNING id, sentence_text, source, created_at;
	`

	var s SentenceResponse
	err := r.pool.QueryRow(ctx, query, flashcardID, text, source, userID).Scan(&s.ID, &s.SentenceText, &s.Source, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SentenceResponse{}, ErrNotFound
		}
		return SentenceResponse{}, fmt.Errorf("failed to add sentence: %w", err)
	}
	return s, nil
}

func (r *pgRepository) SaveGeneratedSentences(ctx context.Context, flashcardID string, userID int64, sentences []string) ([]string, error) {
	var owned bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM flashcards WHERE id = $1 AND user_id = $2)`, flashcardID, userID).Scan(&owned)
	if err != nil {
		return nil, fmt.Errorf("failed to verify flashcard ownership: %w", err)
	}
	if !owned {
		return nil, ErrNotFound
	}

	saved := make([]string, 0, len(sentences))
	for _, sentence := range sentences {
		const query = `
			INSERT INTO ai_suggested_sentences (flashcard_id, sentence_text)
			VALUES ($1, $2)
			RETURNING sentence_text;
		`
		var text string
		if err := r.pool.QueryRow(ctx, query, flashcardID, sentence).Scan(&text); err != nil {
			return nil, fmt.Errorf("failed to save generated sentence: %w", err)
		}
		saved = append(saved, text)
	}

	return saved, nil
}

func (r *pgRepository) DeleteSentence(ctx context.Context, sentenceID int64, userID int64) error {
	const query = `
		DELETE FROM flashcard_sentences fs
		USING flashcards f
		WHERE fs.id = $1 AND fs.flashcard_id = f.id AND f.user_id = $2;
	`
	tag, err := r.pool.Exec(ctx, query, sentenceID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sentence: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
