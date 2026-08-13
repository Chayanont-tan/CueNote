package flashcard

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mission-note/internal/core/middleware"
	"mission-note/internal/core/response"
	"mission-note/internal/core/validation"
)

type handler struct {
	service Service
}

func newHandler(service Service) *handler {
	return &handler{service: service}
}

func respondNotFoundAware(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		response.Error(c, http.StatusNotFound, "not found")
		return
	}
	response.Error(c, http.StatusInternalServerError, err.Error())
}

// --- Tags ---

func (h *handler) createTag(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, validation.BindErrorMessage(err))
		return
	}

	tag, err := h.service.CreateTag(c.Request.Context(), userID, req.Name)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, tag)
}

func (h *handler) listTags(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))

	tags, err := h.service.ListTags(c.Request.Context(), userID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, tags)
}

// --- Flashcards ---

func (h *handler) generateFlashcards(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	tagID, err := strconv.ParseInt(c.Param("tag_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tag id")
		return
	}

	var req GenerateFlashcardsRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.Error(c, http.StatusBadRequest, validation.BindErrorMessage(err))
		return
	}

	result, err := h.service.PreviewFlashcard(c.Request.Context(), tagID, userID, req.Word)
	if err != nil {
		if errors.Is(err, ErrVocabAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		respondNotFoundAware(c, err)
		return
	}
	slog.Info("result", "result", result)
	response.Success(c, http.StatusCreated, result)
}

func (h *handler) saveFlashcard(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req SaveFlashcardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, validation.BindErrorMessage(err))
		return
	}

	result, err := h.service.SaveFlashcard(c.Request.Context(), userID, req)
	if err != nil {
		respondNotFoundAware(c, err)
		return
	}
	response.Success(c, http.StatusCreated, result)
}

func (h *handler) listFlashcards(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	tagID, err := strconv.ParseInt(c.Param("tag_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid tag id")
		return
	}

	cards, err := h.service.ListFlashcards(c.Request.Context(), tagID, userID, c.Query("level"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, cards)
}

func (h *handler) getFlashcard(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	card, err := h.service.GetFlashcard(c.Request.Context(), c.Param("card_id"), userID)
	if err != nil {
		respondNotFoundAware(c, err)
		return
	}
	response.Success(c, http.StatusOK, card)
}

// --- Sentences ---

func (h *handler) addSentence(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AddSentenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, validation.BindErrorMessage(err))
		return
	}

	sentence, err := h.service.AddSentence(c.Request.Context(), c.Param("card_id"), userID, req.SentenceText)
	if err != nil {
		respondNotFoundAware(c, err)
		return
	}
	response.Success(c, http.StatusCreated, sentence)
}

func (h *handler) generateSentences(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	sentences, err := h.service.GenerateSentences(c.Request.Context(), c.Param("card_id"), userID)
	if err != nil {
		respondNotFoundAware(c, err)
		return
	}
	response.Success(c, http.StatusOK, GenerateSentencesResponse{Sentences: sentences})
}

func (h *handler) deleteSentence(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	sentenceID, err := strconv.ParseInt(c.Param("sentence_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid sentence id")
		return
	}

	if err := h.service.DeleteSentence(c.Request.Context(), sentenceID, userID); err != nil {
		respondNotFoundAware(c, err)
		return
	}
	response.Success(c, http.StatusOK, gin.H{"deleted": true})
}
