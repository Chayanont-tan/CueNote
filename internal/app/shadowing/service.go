package shadowing

import (
	"context"
	"encoding/json"
	"io"

	"mission-note/internal/pkg/openai"
)

type service struct {
	repo Repository
	ai   *openai.Client
}

// NewService creates the shadowing Service.
func NewService(repo Repository, ai *openai.Client) Service {
	return &service{repo: repo, ai: ai}
}

func (s *service) GetSentence(ctx context.Context, userID, sentenceID int64) (SentenceResponse, error) {
	sentence, err := s.repo.FindSentenceForUser(ctx, sentenceID, userID)
	if err != nil {
		return SentenceResponse{}, err
	}
	return toSentenceResponse(sentence), nil
}

func (s *service) SubmitAttempt(ctx context.Context, userID int64, req SubmitAttemptRequest) (ScoreResponse, error) {
	sentence, err := s.repo.FindSentenceForUser(ctx, req.SentenceID, userID)
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

	// The audio is never persisted — it's transcribed straight from memory
	// and discarded once we have the text.
	transcript, err := s.ai.TranscribeAudio(ctx, req.Audio.Filename, data)
	if err != nil {
		return ScoreResponse{}, err
	}

	score, correctWords, mispronouncedWords := scoreTranscript(sentence.Text, transcript)

	correctJSON, err := json.Marshal(correctWords)
	if err != nil {
		return ScoreResponse{}, err
	}
	mispronouncedJSON, err := json.Marshal(mispronouncedWords)
	if err != nil {
		return ScoreResponse{}, err
	}

	if _, err := s.repo.SaveAttempt(ctx, Attempt{
		UserID:             userID,
		SentenceID:         req.SentenceID,
		Transcript:         transcript,
		Score:              score,
		CorrectWords:       correctJSON,
		MispronouncedWords: mispronouncedJSON,
	}); err != nil {
		return ScoreResponse{}, err
	}

	return ScoreResponse{
		Score:              score,
		CorrectWords:       correctWords,
		MispronouncedWords: mispronouncedWords,
	}, nil
}
