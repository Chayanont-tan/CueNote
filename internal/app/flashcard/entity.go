package flashcard

import "time"

type Vocabulary struct {
	ID           int64     `db:"id"`
	Word         string    `db:"word"`
	PartOfSpeech string    `db:"part_of_speech"`
	MeaningTH    string    `db:"meaning_th"`
	Level        string    `db:"level"`
	CreatedAt    time.Time `db:"created_at"`
}

type Tag struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedBy int64     `db:"created_by"`
	CreatedAt time.Time `db:"created_at"`
}

type VocabularyTag struct {
	VocabularyID int64 `db:"vocabulary_id"`
	TagID        int64 `db:"tag_id"`
}

type Flashcard struct {
	ID           string    `db:"id"`
	UserID       int64     `db:"user_id"`
	VocabularyID int64     `db:"vocabulary_id"`
	ImageURL     string    `db:"image_url"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type FlashcardSentence struct {
	ID           int64     `db:"id"`
	FlashcardID  string    `db:"flashcard_id"`
	SentenceText string    `db:"sentence_text"`
	Source       string    `db:"source"`
	CreatedAt    time.Time `db:"created_at"`
}

type AISuggestedSentence struct {
	ID           int64     `db:"id"`
	FlashcardID  string    `db:"flashcard_id"`
	SentenceText string    `db:"sentence_text"`
	CreatedAt    time.Time `db:"created_at"`
}
