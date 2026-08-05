package shadowing

import "mime/multipart"

// SubmitAttemptRequest carries the user's recorded shadowing audio for scoring.
type SubmitAttemptRequest struct {
	UserID     int64                 `form:"user_id" binding:"required"`
	SentenceID int64                 `form:"sentence_id" binding:"required"`
	Audio      *multipart.FileHeader `form:"audio" binding:"required"`
}

// SentenceResponse is the API representation of a Sentence.
type SentenceResponse struct {
	ID         int64  `json:"id"`
	Text       string `json:"text"`
	AudioURL   string `json:"audio_url"`
	Timestamps []byte `json:"timestamps"`
}

// ScoreResponse is the result of a shadowing attempt.
type ScoreResponse struct {
	Score              int      `json:"score"`
	CorrectWords       []string `json:"correct_words"`
	MispronouncedWords []string `json:"mispronounced_words"`
}

func toSentenceResponse(s Sentence) SentenceResponse {
	return SentenceResponse{
		ID:         s.ID,
		Text:       s.Text,
		AudioURL:   s.AudioURL,
		Timestamps: s.Timestamps,
	}
}
