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

func NewService(repo Repository, openaiClient *openai.Client) Service {
	return &service{repo: repo, openaiClient: openaiClient}
}

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

func (s *service) PreviewFlashcard(ctx context.Context, tagID int64, userID int64, word string) (*PreviewFlashcardResponse, error) {

	tag, err := s.repo.GetTagByID(ctx, tagID, userID)
	if err != nil {
		return nil, err
	}

	word = strings.TrimSpace(word)

	if word != "" {
		vocab, err := s.resolveUserWord(ctx, word)
		if err != nil {
			return nil, err
		}

		sentences, err := s.openaiClient.GenerateSentences(ctx, vocab.Word, vocab.MeaningTH)
		if err != nil {
			return nil, err
		}

		return &PreviewFlashcardResponse{
			Flashcards: []FlashcardResponse{{
				FlashcardBase: FlashcardBase{
					Word:         vocab.Word,
					PartOfSpeech: vocab.PartOfSpeech,
					MeaningTH:    vocab.MeaningTH,
					Level:        vocab.Level,
				},
				AISuggestedSentences: sentences,
			}},
		}, nil
	}

	existingWords, err := s.repo.ListVocabWordsForTag(ctx, tag.ID)
	if err != nil {
		return nil, err
	}

	aiResponse, err := s.openaiClient.GenerateVocabulariesByTag(ctx, tag.Name, 1, existingWords)
	if err != nil {
		return nil, err
	}

	flashcards := make([]FlashcardResponse, 0, len(aiResponse.Vocabularies))
	for _, item := range aiResponse.Vocabularies {
		sentences, err := s.openaiClient.GenerateSentences(ctx, item.Word, item.MeaningTH)
		if err != nil {
			return nil, err
		}

		flashcards = append(flashcards, FlashcardResponse{
			FlashcardBase: FlashcardBase{
				Word:         item.Word,
				PartOfSpeech: item.PartOfSpeech,
				MeaningTH:    item.MeaningTH,
				Level:        item.Level,
			},
			AISuggestedSentences: sentences,
		})
	}

	return &PreviewFlashcardResponse{Flashcards: flashcards}, nil
}

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

func (s *service) SaveFlashcard(ctx context.Context, userID int64, req SaveFlashcardRequest) (FlashcardResponse, error) {
	tag, err := s.repo.GetTagByID(ctx, req.TagID, userID)
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
		ID: fc.ID,
		FlashcardBase: FlashcardBase{
			Word:         vocab.Word,
			PartOfSpeech: vocab.PartOfSpeech,
			MeaningTH:    vocab.MeaningTH,
			Level:        vocab.Level,
		},
		AISuggestedSentences: aiSentences,
		CreatedAt:            fc.CreatedAt,
	}, nil
}

func (s *service) ListFlashcards(ctx context.Context, tagID int64, userID int64) (FlashcardsForTagResponse, error) {
	tag, err := s.repo.GetTagByID(ctx, tagID, userID)
	if err != nil {
		return FlashcardsForTagResponse{}, err
	}

	cards, err := s.repo.ListFlashcardsForTag(ctx, tagID, userID)
	if err != nil {
		return FlashcardsForTagResponse{}, err
	}

	return FlashcardsForTagResponse{TagName: tag.Name, Flashcards: cards}, nil
}

func (s *service) GetFlashcard(ctx context.Context, flashcardID string, userID int64) (FlashcardResponse, error) {
	return s.repo.GetFlashcardByID(ctx, flashcardID, userID)
}

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
