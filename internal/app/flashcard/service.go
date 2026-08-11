package flashcard

import (
	"context"
	"math"
	"strings"

	"mission-note/internal/pkg/openai"
)

type service struct {
	repo         Repository
	openaiClient *openai.Client
}

// NewService creates the flashcard Service.
func NewService(repo Repository, openaiClient *openai.Client) Service {
	return &service{repo: repo, openaiClient: openaiClient}
}

// --- Tags ---

func (s *service) CreateTag(ctx context.Context, userID int64, name string) (TagResponse, error) {
	tag, err := s.repo.CreateTag(ctx, userID, name)
	if err != nil {
		return TagResponse{}, err
	}
	return TagResponse{ID: tag.ID, Name: tag.Name}, nil
}

func (s *service) ListTags(ctx context.Context, userID int64, limit int) ([]TagResponse, error) {
	if limit <= 0 {
		limit = math.MaxInt32
	}
	tags, err := s.repo.ListTagsForUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// --- Flashcards ---

// GenerateFlashcards ถ้า user พิมพ์คำศัพท์เองมาใน word ก็ใช้คำนั้นตรงๆ (AI เติมแค่
// part_of_speech/meaning_th/level ให้คำที่ยังไม่มีในระบบ) ถ้าไม่ส่ง word มาเลยถึงให้ AI
// เจนคำศัพท์ใหม่ 1 คำจากชื่อ tag แทน (ไม่เอาคำที่ tag นี้มีอยู่แล้วซ้ำ) แล้วสร้าง flashcard
// ใหม่ (ประโยคตัวอย่างชุดใหม่เสมอ แม้คำจะซ้ำเดิม — ไม่เจนรูป ใช้ icon แทนฝั่ง frontend)
func (s *service) GenerateFlashcards(ctx context.Context, tagID int64, userID int64, word string) (*GenerateFlashcardsResponse, error) {
	tag, err := s.repo.GetTagByID(ctx, tagID, userID) // เช็คก่อนว่า tag นี้เป็นของ user คนนี้จริง ก่อนจะเจน flashcard เพิ่มให้
	if err != nil {
		return nil, err
	}

	word = strings.TrimSpace(word)

	var items []Vocabulary
	if word != "" {
		item, err := s.resolveUserWord(ctx, tag, word) // ใช้คำที่ user พิมพ์มาเองตรงๆ (เช็คก่อนว่ามีอยู่ในระบบแล้วหรือยัง ถ้ามีแล้วก็เอา part_of_speech/meaning_th/level เดิมมาใช้เลย)
		if err != nil {
			return nil, err
		}
		items = []Vocabulary{item}
	} else {
		items, err = s.generateVocabFromTag(ctx, tag) // ดูจาก tag ก่อนว่าอะไร ละเช็ค ว่ามีคำไหนใน tag ละบ้างจะได้ไม่ซ้ำ
		if err != nil {
			return nil, err
		}
	}

	flashcards := make([]GeneratedFlashcardResponse, 0, len(items))
	for _, item := range items {
		sentences, err := s.openaiClient.GenerateSentences(ctx, item.Word, item.MeaningTH)
		if err != nil {
			return nil, err
		}

		fc, savedSentences, err := s.repo.CreateFlashcardWithSentences(ctx, userID, item.ID, "", sentences)
		if err != nil {
			return nil, err
		}

		aiSentences := make([]string, 0, len(savedSentences))
		for _, saved := range savedSentences {
			aiSentences = append(aiSentences, saved.SentenceText)
		}

		flashcards = append(flashcards, GeneratedFlashcardResponse{
			ID:                   fc.ID,
			Word:                 item.Word,
			PartOfSpeech:         item.PartOfSpeech,
			MeaningTH:            item.MeaningTH,
			AISuggestedSentences: aiSentences,
			CreatedAt:            fc.CreatedAt,
		})
	}

	return &GenerateFlashcardsResponse{Flashcards: flashcards}, nil
}

// resolveUserWord ใช้คำที่ user พิมพ์เองตรงๆ — เช็คก่อนว่ามีอยู่ในระบบแล้วหรือยัง (ไม่งั้นจะ
// เจน AI ซ้ำโดยไม่จำเป็น) มีแล้วก็เอา part_of_speech/meaning_th/level เดิมมาใช้เลย ถ้ายังไม่มี
// ค่อยให้ AI เติมรายละเอียดให้คำนั้น
func (s *service) resolveUserWord(ctx context.Context, tag TagResponse, word string) (Vocabulary, error) {
	var item openai.OpenAIVocabItem

	existing, found, err := s.repo.FindVocabByWord(ctx, word) // หาในระบบก่อนว่ามีคำนี้อยู่แล้วหรือยัง ถ้ามีแล้วก็เอา part_of_speech/meaning_th/level เดิมมาใช้เลย
	if err != nil {
		return Vocabulary{}, err
	}
	if found {
		item = openai.OpenAIVocabItem{
			Word:         existing.Word,
			PartOfSpeech: existing.PartOfSpeech,
			MeaningTH:    existing.MeaningTH,
			Level:        existing.Level,
		}
	} else {
		detail, err := s.openaiClient.GenerateVocabularyDetails(ctx, word)
		if err != nil {
			return Vocabulary{}, err
		}
		item = *detail
	}

	saved, err := s.repo.SaveVocabulariesForTag(ctx, tag.ID, []openai.OpenAIVocabItem{item})
	if err != nil {
		return Vocabulary{}, err
	}
	return saved[0], nil
}

// generateVocabFromTag ไม่มี user พิมพ์คำมา ให้ AI เจนคำศัพท์ใหม่ 1 คำจากชื่อ tag เสมอ โดยบอก
// AI ไม่ให้เจนคำที่ tag นี้มีอยู่แล้วซ้ำ
func (s *service) generateVocabFromTag(ctx context.Context, tag TagResponse) ([]Vocabulary, error) {
	existingWords, err := s.repo.ListVocabWordsForTag(ctx, tag.ID)
	if err != nil {
		return nil, err
	}

	aiResponse, err := s.openaiClient.GenerateVocabulariesByTag(ctx, tag.Name, 1, existingWords)
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
