package shadowing

import "mime/multipart"

// SubmitAttemptRequest carries the user's recorded shadowing audio for scoring.
// UserID is not part of the request body — it comes from the JWT.
type SubmitAttemptRequest struct {
	SentenceID int64                 `form:"sentence_id" binding:"required"`
	Audio      *multipart.FileHeader `form:"audio" binding:"required"`
}

// SentenceResponse is the API representation of a sentence to shadow-read.
type SentenceResponse struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

// ScoreResponse is the result of a shadowing attempt.
type ScoreResponse struct {
	Score              int      `json:"score"`
	CorrectWords       []string `json:"correct_words"`
	MispronouncedWords []string `json:"mispronounced_words"`
}

func toSentenceResponse(s SentenceRef) SentenceResponse {
	return SentenceResponse{
		ID:   s.ID,
		Text: s.Text,
	}
}
