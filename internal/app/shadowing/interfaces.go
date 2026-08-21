package shadowing

import "context"

// Repository defines persistence operations for the shadowing feature.
type Repository interface {
	FindSentenceForUser(ctx context.Context, sentenceID, userID int64) (SentenceRef, error)
	SaveAttempt(ctx context.Context, a Attempt) (Attempt, error)
}

// Service defines the business logic operations for the shadowing feature.
type Service interface {
	GetSentence(ctx context.Context, userID, sentenceID int64) (SentenceResponse, error)
	SubmitAttempt(ctx context.Context, userID int64, req SubmitAttemptRequest) (ScoreResponse, error)
}
