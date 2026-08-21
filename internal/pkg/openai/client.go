package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	openaiSDK "github.com/sashabaranov/go-openai"
)

// Client wraps the OpenAI/Groq SDK for the features this app needs.
type Client struct {
	apiKey     string
	model      string
	sttModel   string
	mockImages bool
	sdk        *openaiSDK.Client
}

// TranscribeAudio sends raw audio bytes to the configured speech-to-text
// model and returns the transcribed text. The audio is streamed straight
// from memory — nothing is written to disk.
func (c *Client) TranscribeAudio(ctx context.Context, filename string, data []byte) (string, error) {
	resp, err := c.sdk.CreateTranscription(ctx, openaiSDK.AudioRequest{
		Model:    c.sttModel,
		Reader:   bytes.NewReader(data),
		FilePath: filename,
		Format:   openaiSDK.AudioResponseFormatText,
	})
	if err != nil {
		return "", fmt.Errorf("ai transcription failed: %w", err)
	}
	return resp.Text, nil
}

// New creates a new LLM client supporting both OpenAI and Groq (via baseURL).
// mockImages, when true, makes GenerateImage return a placeholder URL instead
// of calling the (paid) DALL-E endpoint.
func New(apiKey, baseURL, model, sttModel string, mockImages bool) *Client {
	config := openaiSDK.DefaultConfig(apiKey)

	// ถ้ามีการระบุ BaseURL (เช่น Groq: https://api.groq.com/openai/v1) ให้ override เข้าไป
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	// ถ้าไม่ระบุ Model ให้ใช้ gpt-4o-mini เป็นค่า Default
	selectedModel := model
	if selectedModel == "" {
		selectedModel = openaiSDK.GPT4oMini
	}

	selectedSTTModel := sttModel
	if selectedSTTModel == "" {
		selectedSTTModel = openaiSDK.Whisper1
	}

	return &Client{
		apiKey:     apiKey,
		model:      selectedModel,
		sttModel:   selectedSTTModel,
		mockImages: mockImages,
		sdk:        openaiSDK.NewClientWithConfig(config),
	}
}

// OpenAIVocabGenResponse คือโครงสร้าง JSON ที่ต้องการให้ AI ตอบกลับมา
type OpenAIVocabGenResponse struct {
	TagName      string            `json:"tag_name"`
	Vocabularies []OpenAIVocabItem `json:"vocabularies"`
}

type OpenAIVocabItem struct {
	Word         string   `json:"word"`
	PartOfSpeech string   `json:"part_of_speech"`
	MeaningTH    string   `json:"meaning_th"`
	Level        string   `json:"level"`
	AISentences  []string `json:"ai_sentences"`
}

// GenerateVocabulariesByTag ยิง AI เพื่อเจนคำศัพท์และประโยคตัวอย่างตาม Tag
// excludeWords (ถ้ามี) คือคำที่ tag นี้มีอยู่แล้ว บอก AI ไว้ไม่ให้เจนซ้ำ
func (c *Client) GenerateVocabulariesByTag(ctx context.Context, tag string, limit int, excludeWords []string) (*OpenAIVocabGenResponse, error) {
	systemPrompt := `You are an English language learning assistant.
Generate English vocabulary words based on the user's requested tag/topic.
You MUST respond strictly in valid JSON format matching this structure:
{
  "tag_name": "string",
  "vocabularies": [
    {
      "word": "string",
      "part_of_speech": "noun/verb/adjective/etc.",
      "meaning_th": "คำแปลภาษาไทย",
      "level": "CEFR level: one of A1, A2, B1, B2, C1, C2 based on word difficulty",
      "ai_sentences": ["example sentence 1", "example sentence 2", "example sentence 3"]
    }
  ]
}`

	userPrompt := fmt.Sprintf("Generate %d vocabulary words for the topic/tag: '%s'. Provide 3 example sentences for each word.", limit, tag)
	if len(excludeWords) > 0 {
		userPrompt += fmt.Sprintf(" Do not repeat these words that are already used: %s.", strings.Join(excludeWords, ", "))
	}

	req := openaiSDK.ChatCompletionRequest{
		Model: c.model, // 👈 ใช้ model ตามที่ตั้งไว้ใน struct
		ResponseFormat: &openaiSDK.ChatCompletionResponseFormat{
			Type: openaiSDK.ChatCompletionResponseFormatTypeJSONObject,
		},
		Messages: []openaiSDK.ChatCompletionMessage{
			{
				Role:    openaiSDK.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openaiSDK.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.7,
	}

	resp, err := c.sdk.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ai chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from ai")
	}

	var result OpenAIVocabGenResponse
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ai response: %w", err)
	}

	return &result, nil
}

// GenerateVocabularyDetails ยิง AI เพื่อเติม part_of_speech/meaning_th/level ให้คำที่ user
// พิมพ์เอง (คำนี้ยังไม่มีอยู่ในระบบมาก่อน)
func (c *Client) GenerateVocabularyDetails(ctx context.Context, word string) (*OpenAIVocabItem, error) {
	systemPrompt := `You are an English language learning assistant.
You MUST respond strictly in valid JSON format matching this structure:
{
  "word": "string",
  "part_of_speech": "noun/verb/adjective/etc.",
  "meaning_th": "คำแปลภาษาไทย",
  "level": "CEFR level: one of A1, A2, B1, B2, C1, C2 based on word difficulty"
}`

	userPrompt := fmt.Sprintf("Give the part of speech, Thai meaning, and CEFR level for the English word '%s'.", word)

	req := openaiSDK.ChatCompletionRequest{
		Model: c.model,
		ResponseFormat: &openaiSDK.ChatCompletionResponseFormat{
			Type: openaiSDK.ChatCompletionResponseFormatTypeJSONObject,
		},
		Messages: []openaiSDK.ChatCompletionMessage{
			{
				Role:    openaiSDK.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openaiSDK.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.7,
	}

	resp, err := c.sdk.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ai chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from ai")
	}

	var result OpenAIVocabItem
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ai response: %w", err)
	}

	return &result, nil
}

// GenerateImage เจนรูปประกอบคำศัพท์ด้วย DALL-E ถ้า mockImages เป็น true จะคืน
// placeholder URL ทันทีโดยไม่เรียก API จริง (ไว้ใช้ตอน dev เพื่อไม่ให้เสียเงิน)
func (c *Client) GenerateImage(ctx context.Context, prompt string) (string, error) {
	if c.mockImages {
		return "https://placehold.co/512x512?text=" + url.QueryEscape(prompt), nil
	}

	resp, err := c.sdk.CreateImage(ctx, openaiSDK.ImageRequest{
		Prompt:         prompt,
		Model:          openaiSDK.CreateImageModelDallE3,
		N:              1,
		Size:           openaiSDK.CreateImageSize1024x1024,
		ResponseFormat: openaiSDK.CreateImageResponseFormatURL,
	})
	if err != nil {
		return "", fmt.Errorf("ai image generation failed: %w", err)
	}

	if len(resp.Data) == 0 {
		return "", fmt.Errorf("empty image response from ai")
	}

	return resp.Data[0].URL, nil
}

// OpenAISentencesResponse คือโครงสร้าง JSON ที่ต้องการให้ AI ตอบกลับมาสำหรับ GenerateSentences
type OpenAISentencesResponse struct {
	Sentences []string `json:"sentences"`
}

// GenerateSentences ยิง AI เพื่อเจนประโยคตัวอย่างใหม่สำหรับคำศัพท์คำเดียว
// ใช้ตอนสร้าง flashcard แต่ละใบ เพื่อให้คำซ้ำก็ยังได้ประโยคใหม่ทุกครั้ง
func (c *Client) GenerateSentences(ctx context.Context, word, meaningTH string) ([]string, error) {
	systemPrompt := `You are an English language learning assistant.
You MUST respond strictly in valid JSON format matching this structure:
{
  "sentences": ["example sentence 1"]
}`

	userPrompt := fmt.Sprintf("Generate 1 example sentence using the English word '%s' (Thai meaning: %s).", word, meaningTH)

	req := openaiSDK.ChatCompletionRequest{
		Model: c.model,
		ResponseFormat: &openaiSDK.ChatCompletionResponseFormat{
			Type: openaiSDK.ChatCompletionResponseFormatTypeJSONObject,
		},
		Messages: []openaiSDK.ChatCompletionMessage{
			{
				Role:    openaiSDK.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openaiSDK.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.7,
	}

	resp, err := c.sdk.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ai chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from ai")
	}

	var result OpenAISentencesResponse
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ai response: %w", err)
	}

	return result.Sentences, nil
}
