package flashcard

import (
	"context"

	"mission-note/internal/pkg/openai"
)

// Repository defines persistence operations for the flashcard feature —
// covers tags, the vocabulary catalog, flashcards, and their sentences.
type Repository interface {
	// Tags
	CreateTag(ctx context.Context, userID int64, name string) (Tag, error)
	ListTagsForUser(ctx context.Context, userID int64, limit int) ([]TagResponse, error)
	GetTagByID(ctx context.Context, tagID int64, userID int64) (TagResponse, error)

	// Vocabulary catalog (internal to this module — no HTTP endpoint of its own)
	FindVocabByWord(ctx context.Context, word string) (Vocabulary, bool, error)
	ListVocabWordsForTag(ctx context.Context, tagID int64) ([]string, error)
	SaveVocabulariesForTag(ctx context.Context, tagID int64, items []openai.OpenAIVocabItem) ([]Vocabulary, error)

	// Flashcards
	CreateFlashcardWithSentences(ctx context.Context, userID int64, vocabularyID int64, imageURL string, aiSentences []string) (Flashcard, []AISuggestedSentence, error)
	ListFlashcardsForTag(ctx context.Context, tagID int64, userID int64, level string) ([]FlashcardResponse, error)
	GetFlashcardByID(ctx context.Context, flashcardID string, userID int64) (FlashcardResponse, error)

	// Sentences
	AddSentence(ctx context.Context, flashcardID string, userID int64, text string, source string) (SentenceResponse, error)
	SaveGeneratedSentences(ctx context.Context, flashcardID string, userID int64, sentences []string) ([]string, error)
	DeleteSentence(ctx context.Context, sentenceID int64, userID int64) error
}

// Service defines the business logic operations for the flashcard feature.
type Service interface {
	CreateTag(ctx context.Context, userID int64, name string) (TagResponse, error)
	ListTags(ctx context.Context, userID int64, limit int) ([]TagResponse, error)

	GenerateFlashcards(ctx context.Context, tagID int64, userID int64, word string) (*GenerateFlashcardsResponse, error)
	ListFlashcards(ctx context.Context, tagID int64, userID int64, level string) ([]FlashcardResponse, error)
	GetFlashcard(ctx context.Context, flashcardID string, userID int64) (FlashcardResponse, error)

	AddSentence(ctx context.Context, flashcardID string, userID int64, text string) (SentenceResponse, error)
	GenerateSentences(ctx context.Context, flashcardID string, userID int64) ([]string, error)
	DeleteSentence(ctx context.Context, sentenceID int64, userID int64) error
}
