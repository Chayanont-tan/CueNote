package shadowing

import "context"

// Repository defines persistence operations for the shadowing feature.
type Repository interface {
	FindSentenceByID(ctx context.Context, id int64) (Sentence, error)
	SaveAttempt(ctx context.Context, a Attempt) (Attempt, error)
}

// Service defines the business logic operations for the shadowing feature.
type Service interface {
	GetSentence(ctx context.Context, id int64) (SentenceResponse, error)
	SubmitAttempt(ctx context.Context, req SubmitAttemptRequest) (ScoreResponse, error)
}
