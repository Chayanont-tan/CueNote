package shadowing

import (
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

func (h *handler) getSentence(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid sentence id")
		return
	}

	sentence, err := h.service.GetSentence(c.Request.Context(), userID, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, sentence)
}

func (h *handler) submitAttempt(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req SubmitAttemptRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, validation.BindErrorMessage(err))
		return
	}

	score, err := h.service.SubmitAttempt(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, score)
}
