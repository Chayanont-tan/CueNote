package flashcard

import "time"

// --- Tags ---

type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

type TagResponse struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	IconURL      *string `json:"icon_url"`
	BadgeIconURL *string `json:"badge_icon_url"`
	TotalWords   int     `json:"total_words"`
}

// --- Flashcards ---

type GenerateFlashcardsRequest struct {
	Word string `json:"word"`
}

type SentenceResponse struct {
	ID           int64     `json:"id"`
	SentenceText string    `json:"sentence_text"`
	Source       string    `json:"source"`
	CreatedAt    time.Time `json:"created_at"`
}

type FlashcardResponse struct {
	ID                   string             `json:"id"`
	Word                 string             `json:"word"`
	PartOfSpeech         string             `json:"part_of_speech"`
	MeaningTH            string             `json:"meaning_th"`
	Level                string             `json:"level"`
	ImageURL             string             `json:"image_url"`
	AISuggestedSentences []string           `json:"ai_suggested_sentences"`
	Sentences            []SentenceResponse `json:"sentences"`
	CreatedAt            time.Time          `json:"created_at"`
}

type PreviewFlashcardItem struct {
	ID                   string    `json:"id"`
	Word                 string    `json:"word"`
	PartOfSpeech         string    `json:"part_of_speech"`
	MeaningTH            string    `json:"meaning_th"`
	AISuggestedSentences []string  `json:"ai_suggested_sentences"`
	CreatedAt            time.Time `json:"created_at"`
}

type PreviewFlashcardResponse struct {
	Flashcards []PreviewFlashcardItem `json:"flashcards"`
}

// SaveFlashcardRequest คือคำศัพท์ที่ผ่านการ preview มาแล้ว (ทั้งจาก path user พิมพ์เอง
// และ path AI เจนจาก tag) ให้ frontend ส่งกลับมาตอนกด save จริง
type SaveFlashcardRequest struct {
	TagID        int64    `json:"tag_id" binding:"required"`
	Word         string   `json:"word" binding:"required"`
	PartOfSpeech string   `json:"part_of_speech"`
	MeaningTH    string   `json:"meaning_th" binding:"required"`
	Level        string   `json:"level"`
	Sentences    []string `json:"sentences"`
}

// --- Sentences ---

type AddSentenceRequest struct {
	SentenceText string `json:"sentence_text" binding:"required"`
}

type GenerateSentencesResponse struct {
	Sentences []string `json:"sentences"`
}
