package flashcard

import (
	"context"
	"fmt"

	"mission-note/internal/pkg/openai"
)

// unlockThresholds คือเกณฑ์ปลดล็อก Shadowing scenario ทีละขั้น (scenario 1/2/3)
var unlockThresholds = []int{5, 10, 15}

// nextUnlockThreshold คืนเกณฑ์คำศัพท์ถัดไปที่ยังไม่ปลดล็อก ถ้าปลดล็อกครบทุกขั้นแล้ว
// คืนเกณฑ์สูงสุด (ไม่มี scenario เพิ่มให้ปลดล็อกอีกในตอนนี้)
func nextUnlockThreshold(unlockedWords int) int {
	for _, threshold := range unlockThresholds {
		if unlockedWords < threshold {
			return threshold
		}
	}
	return unlockThresholds[len(unlockThresholds)-1]
}

type service struct {
	repo         Repository
	openaiClient *openai.Client
}

// NewService creates the flashcard Service.
func NewService(repo Repository, openaiClient *openai.Client) Service {
	return &service{repo: repo, openaiClient: openaiClient}
}

// --- Tags ---

func (s *service) CreateTag(ctx context.Context, name string) (TagResponse, error) {
	tag, err := s.repo.CreateTag(ctx, name)
	if err != nil {
		return TagResponse{}, err
	}
	return TagResponse{ID: tag.ID, Name: tag.Name, UnlockThreshold: nextUnlockThreshold(0)}, nil
}

func (s *service) ListTags(ctx context.Context, userID int64, limit int) ([]TagResponse, error) {
	tags, err := s.repo.ListTagsForUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	for i := range tags {
		tags[i].UnlockThreshold = nextUnlockThreshold(tags[i].UnlockedWords)
	}
	return tags, nil
}

func (s *service) GetTag(ctx context.Context, tagID int64, userID int64) (TagResponse, error) {
	tag, err := s.repo.GetTagByID(ctx, tagID, userID)
	if err != nil {
		return TagResponse{}, err
	}
	tag.UnlockThreshold = nextUnlockThreshold(tag.UnlockedWords)
	return tag, nil
}

// --- Flashcards ---

// GenerateFlashcards ค้นหาคำศัพท์ที่มีอยู่แล้วใน tag นี้ ถ้าไม่มีจะเจนใหม่ด้วย AI
// แล้วต่อคำสร้าง flashcard ใหม่ทุกใบ (รูป + ประโยคตัวอย่างชุดใหม่เสมอ แม้คำจะซ้ำเดิม)
func (s *service) GenerateFlashcards(ctx context.Context, tagID int64, userID int64, limit int) (*GenerateFlashcardsResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	tag, err := s.repo.GetTagByID(ctx, tagID, userID)
	if err != nil {
		return nil, err
	}

	items, err := s.getOrGenerateVocab(ctx, tag, limit)
	if err != nil {
		return nil, err
	}

	flashcards := make([]FlashcardResponse, 0, len(items))
	for _, item := range items {
		prompt := fmt.Sprintf("An illustration representing the English word '%s' (%s): %s", item.Word, item.PartOfSpeech, item.MeaningTH)
		imageURL, err := s.openaiClient.GenerateImage(ctx, prompt)
		if err != nil {
			return nil, err
		}

		sentences, err := s.openaiClient.GenerateSentences(ctx, item.Word, item.MeaningTH)
		if err != nil {
			return nil, err
		}

		fc, savedSentences, err := s.repo.CreateFlashcardWithSentences(ctx, userID, item.ID, imageURL, sentences)
		if err != nil {
			return nil, err
		}

		aiSentences := make([]string, 0, len(savedSentences))
		for _, saved := range savedSentences {
			aiSentences = append(aiSentences, saved.SentenceText)
		}

		flashcards = append(flashcards, FlashcardResponse{
			ID:                   fc.ID,
			Word:                 item.Word,
			PartOfSpeech:         item.PartOfSpeech,
			MeaningTH:            item.MeaningTH,
			Level:                item.Level,
			ImageURL:             fc.ImageURL,
			AISuggestedSentences: aiSentences,
			CreatedAt:            fc.CreatedAt,
		})
	}

	// เช็ค progress ล่าสุดหลังสร้าง flashcard ใหม่ไปแล้ว (unlocked_words เปลี่ยนไปแล้ว)
	refreshedTag, err := s.repo.GetTagByID(ctx, tagID, userID)
	if err != nil {
		return nil, err
	}
	refreshedTag.UnlockThreshold = nextUnlockThreshold(refreshedTag.UnlockedWords)

	return &GenerateFlashcardsResponse{
		Tag:        refreshedTag,
		Flashcards: flashcards,
	}, nil
}

// getOrGenerateVocab ค้นหาคำศัพท์ที่มีอยู่แล้วใน tag นี้ ถ้าไม่มีจะเจนใหม่ด้วย AI แล้วบันทึกลง DB
func (s *service) getOrGenerateVocab(ctx context.Context, tag TagResponse, limit int) ([]Vocabulary, error) {
	items, err := s.repo.GetRandomVocabByTagID(ctx, tag.ID, limit)
	if err != nil {
		return nil, err
	}

	if len(items) > 0 {
		return items, nil
	}

	aiResponse, err := s.openaiClient.GenerateVocabulariesByTag(ctx, tag.Name, limit)
	if err != nil {
		return nil, err
	}

	return s.repo.SaveVocabulariesForTag(ctx, tag.ID, aiResponse.Vocabularies)
}

func (s *service) ListFlashcards(ctx context.Context, tagID int64, userID int64, level string) ([]FlashcardResponse, error) {
	return s.repo.ListFlashcardsForTag(ctx, tagID, userID, level)
}

func (s *service) GetFlashcard(ctx context.Context, flashcardID string, userID int64) (FlashcardResponse, error) {
	return s.repo.GetFlashcardByID(ctx, flashcardID, userID)
}

// --- Sentences ---

func (s *service) AddSentence(ctx context.Context, flashcardID string, userID int64, text string) (SentenceResponse, error) {
	return s.repo.AddSentence(ctx, flashcardID, userID, text, "user")
}

func (s *service) GenerateSentences(ctx context.Context, flashcardID string, userID int64) ([]string, error) {
	card, err := s.repo.GetFlashcardByID(ctx, flashcardID, userID)
	if err != nil {
		return nil, err
	}

	sentences, err := s.openaiClient.GenerateSentences(ctx, card.Word, card.MeaningTH)
	if err != nil {
		return nil, err
	}

	return s.repo.SaveGeneratedSentences(ctx, flashcardID, userID, sentences)
}

func (s *service) DeleteSentence(ctx context.Context, sentenceID int64, userID int64) error {
	return s.repo.DeleteSentence(ctx, sentenceID, userID)
}
