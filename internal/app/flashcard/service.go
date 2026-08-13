package flashcard

import (
	"context"
	"errors"
	"math"
	"strings"

	"mission-note/internal/pkg/openai"
)

type service struct {
	repo         Repository
	openaiClient *openai.Client
}

var ErrVocabAlreadyExists = errors.New("vocabulary already exists")

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

// PreviewFlashcard ถ้า user พิมพ์คำศัพท์เองมาใน word ก็ใช้คำนั้นตรงๆ (AI เติมแค่
// part_of_speech/meaning_th/level ให้คำที่ยังไม่มีในระบบ) ถ้าไม่ส่ง word มาเลยถึงให้ AI
// เจนคำศัพท์ใหม่ 1 คำจากชื่อ tag แทน (ไม่เอาคำที่ tag นี้มีอยู่แล้วซ้ำ) แล้วสร้าง flashcard
// ใหม่ (ประโยคตัวอย่างชุดใหม่เสมอ แม้คำจะซ้ำเดิม — ไม่เจนรูป ใช้ icon แทนฝั่ง frontend)
func (s *service) PreviewFlashcard(ctx context.Context, tagID int64, userID int64, word string) (*PreviewFlashcardResponse, error) {

	tag, err := s.repo.GetTagByID(ctx, tagID, userID) // เช็คก่อนว่า tag นี้เป็นของ user คนนี้จริง ก่อนจะเจน flashcard เพิ่มให้
	if err != nil {
		return nil, err
	}

	word = strings.TrimSpace(word)

	if word != "" {
		vocab, err := s.resolveUserWord(ctx, word) // ใช้คำที่ user พิมพ์มาเองตรงๆ (เช็คก่อนว่ามีอยู่ในระบบแล้วหรือยัง คำซ้ำจะ error ออกไปเลย)
		if err != nil {
			return nil, err
		}

		sentences, err := s.openaiClient.GenerateSentences(ctx, vocab.Word, vocab.MeaningTH)
		if err != nil {
			return nil, err
		}

		return &PreviewFlashcardResponse{
			Flashcards: []PreviewFlashcardItem{{
				Word:                 vocab.Word,
				PartOfSpeech:         vocab.PartOfSpeech,
				MeaningTH:            vocab.MeaningTH,
				AISuggestedSentences: sentences,
				// ID และ CreatedAt ปล่อย zero value ไว้ — ยังไม่มี flashcard จริงจนกว่าจะเรียก save API แยก
			}},
		}, nil
	}

	items, err := s.generateVocabFromTag(ctx, tag)
	if err != nil {
		return nil, err
	}

	flashcards := make([]PreviewFlashcardItem, 0, len(items))
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

		flashcards = append(flashcards, PreviewFlashcardItem{
			ID:                   fc.ID,
			Word:                 item.Word,
			PartOfSpeech:         item.PartOfSpeech,
			MeaningTH:            item.MeaningTH,
			AISuggestedSentences: aiSentences,
			CreatedAt:            fc.CreatedAt,
		})
	}

	return &PreviewFlashcardResponse{Flashcards: flashcards}, nil
}

// resolveUserWord ใช้คำที่ user พิมพ์เองตรงๆ — เช็คก่อนว่ามีอยู่ในระบบแล้วหรือยัง (ไม่งั้นจะ
// เจน AI ซ้ำโดยไม่จำเป็น) มีแล้วก็เอา part_of_speech/meaning_th/level เดิมมาใช้เลย ถ้ายังไม่มี
// ค่อยให้ AI เติมรายละเอียดให้คำนั้น
func (s *service) resolveUserWord(ctx context.Context, word string) (Vocabulary, error) {
	_, found, err := s.repo.FindVocabByWord(ctx, word)
	if err != nil {
		return Vocabulary{}, err
	}
	if found {
		return Vocabulary{}, ErrVocabAlreadyExists
	} else {
		detail, err := s.openaiClient.GenerateVocabularyDetails(ctx, word)
		if err != nil {
			return Vocabulary{}, err
		}

		vocab := Vocabulary{
			Word:         detail.Word,
			PartOfSpeech: detail.PartOfSpeech,
			MeaningTH:    detail.MeaningTH,
			Level:        detail.Level,
		}
		return vocab, nil
	}
}

// SaveFlashcard บันทึกคำศัพท์ที่ผ่านการ preview มาแล้ว (จาก PreviewFlashcard) ลง DB จริง —
// upsert คำศัพท์เข้า tag ที่ระบุ แล้วสร้าง flashcard พร้อมประโยคตัวอย่างที่ preview มา
func (s *service) SaveFlashcard(ctx context.Context, userID int64, req SaveFlashcardRequest) (FlashcardResponse, error) {
	tag, err := s.repo.GetTagByID(ctx, req.TagID, userID) // เช็คก่อนว่า tag นี้เป็นของ user คนนี้จริง
	if err != nil {
		return FlashcardResponse{}, err
	}

	vocabs, err := s.repo.SaveVocabulariesForTag(ctx, tag.ID, []openai.OpenAIVocabItem{{
		Word:         req.Word,
		PartOfSpeech: req.PartOfSpeech,
		MeaningTH:    req.MeaningTH,
		Level:        req.Level,
	}})
	if err != nil {
		return FlashcardResponse{}, err
	}
	vocab := vocabs[0]

	fc, savedSentences, err := s.repo.CreateFlashcardWithSentences(ctx, userID, vocab.ID, "", req.Sentences)
	if err != nil {
		return FlashcardResponse{}, err
	}

	aiSentences := make([]string, 0, len(savedSentences))
	for _, saved := range savedSentences {
		aiSentences = append(aiSentences, saved.SentenceText)
	}

	return FlashcardResponse{
		ID:                   fc.ID,
		Word:                 vocab.Word,
		PartOfSpeech:         vocab.PartOfSpeech,
		MeaningTH:            vocab.MeaningTH,
		Level:                vocab.Level,
		ImageURL:             fc.ImageURL,
		AISuggestedSentences: aiSentences,
		CreatedAt:            fc.CreatedAt,
	}, nil
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
