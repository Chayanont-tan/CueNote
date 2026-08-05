package shadowing

import (
	"context"
	"io"

	"mission-note/internal/infra/storage"
	"mission-note/internal/pkg/openai"
)

type service struct {
	repo    Repository
	storage storage.Storage
	ai      *openai.Client
}

// NewService creates the shadowing Service.
func NewService(repo Repository, storage storage.Storage, ai *openai.Client) Service {
	return &service{repo: repo, storage: storage, ai: ai}
}

func (s *service) GetSentence(ctx context.Context, id int64) (SentenceResponse, error) {
	sentence, err := s.repo.FindSentenceByID(ctx, id)
	if err != nil {
		return SentenceResponse{}, err
	}
	return toSentenceResponse(sentence), nil
}

func (s *service) SubmitAttempt(ctx context.Context, req SubmitAttemptRequest) (ScoreResponse, error) {
	sentence, err := s.repo.FindSentenceByID(ctx, req.SentenceID)
	if err != nil {
		return ScoreResponse{}, err
	}

	file, err := req.Audio.Open()
	if err != nil {
		return ScoreResponse{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return ScoreResponse{}, err
	}

	audioURL, err := s.storage.Upload(ctx, req.Audio.Filename, data, req.Audio.Header.Get("Content-Type"))
	if err != nil {
		return ScoreResponse{}, err
	}

	transcript, err := s.ai.TranscribeAudio(ctx, data)
	if err != nil {
		return ScoreResponse{}, err
	}

	// TODO: diff-match transcript against sentence.Text to compute a real score.
	_ = transcript
	_ = sentence
	score := 0

	if _, err := s.repo.SaveAttempt(ctx, Attempt{
		UserID:     req.UserID,
		SentenceID: req.SentenceID,
		AudioURL:   audioURL,
		Score:      score,
	}); err != nil {
		return ScoreResponse{}, err
	}

	return ScoreResponse{Score: score}, nil
}
