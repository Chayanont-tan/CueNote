package response

import "github.com/gin-gonic/gin"

// Envelope is the standard JSON response shape for all API endpoints.
type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Success writes a 2xx JSON response with the given payload.
func Success(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{Success: true, Data: data})
}

// Error writes an error JSON response with the given status and message.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Envelope{Success: false, Message: message})
}
