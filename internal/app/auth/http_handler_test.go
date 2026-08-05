package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mission-note/internal/core/response"
)

type mockService struct {
	registerFn func(ctx context.Context, req RegisterRequest) (TokenResponse, error)
	loginFn    func(ctx context.Context, req LoginRequest) (TokenResponse, error)
}

func (m *mockService) Register(ctx context.Context, req RegisterRequest) (TokenResponse, error) {
	return m.registerFn(ctx, req)
}

func (m *mockService) Login(ctx context.Context, req LoginRequest) (TokenResponse, error) {
	return m.loginFn(ctx, req)
}

func newTestRouter(svc Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := newHandler(svc)
	r := gin.New()
	r.POST("/register", h.register)
	r.POST("/login", h.login)
	return r
}

func doRequest(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var env response.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
	return env
}

func TestHandler_Register_Success(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req RegisterRequest) (TokenResponse, error) {
			return TokenResponse{AccessToken: "token-123"}, nil
		},
	}
	r := newTestRouter(svc)

	rec := doRequest(r, http.MethodPost, "/register", RegisterRequest{
		Email:    "user@example.com",
		Password: "correcthorse",
	})

	assert.Equal(t, http.StatusCreated, rec.Code)
	env := decodeEnvelope(t, rec)
	assert.True(t, env.Success)

	data, ok := env.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "token-123", data["access_token"])
}

func TestHandler_Register_InvalidBody(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req RegisterRequest) (TokenResponse, error) {
			t.Fatal("service should not be called for an invalid request body")
			return TokenResponse{}, nil
		},
	}
	r := newTestRouter(svc)

	rec := doRequest(r, http.MethodPost, "/register", RegisterRequest{
		Email:    "not-an-email",
		Password: "short",
	})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	env := decodeEnvelope(t, rec)
	assert.False(t, env.Success)
}

func TestHandler_Register_ServiceError(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req RegisterRequest) (TokenResponse, error) {
			return TokenResponse{}, errors.New("boom")
		},
	}
	r := newTestRouter(svc)

	rec := doRequest(r, http.MethodPost, "/register", RegisterRequest{
		Email:    "user@example.com",
		Password: "correcthorse",
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	env := decodeEnvelope(t, rec)
	assert.False(t, env.Success)
}

func TestHandler_Login_Success(t *testing.T) {
	svc := &mockService{
		loginFn: func(ctx context.Context, req LoginRequest) (TokenResponse, error) {
			return TokenResponse{AccessToken: "token-456"}, nil
		},
	}
	r := newTestRouter(svc)

	rec := doRequest(r, http.MethodPost, "/login", LoginRequest{
		Email:    "user@example.com",
		Password: "correcthorse",
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	env := decodeEnvelope(t, rec)
	assert.True(t, env.Success)

	data, ok := env.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "token-456", data["access_token"])
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	svc := &mockService{
		loginFn: func(ctx context.Context, req LoginRequest) (TokenResponse, error) {
			return TokenResponse{}, ErrInvalidCredentials
		},
	}
	r := newTestRouter(svc)

	rec := doRequest(r, http.MethodPost, "/login", LoginRequest{
		Email:    "user@example.com",
		Password: "wrongpassword",
	})

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	env := decodeEnvelope(t, rec)
	assert.False(t, env.Success)
}
