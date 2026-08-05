package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockRepository struct {
	createFn      func(ctx context.Context, u User) (User, error)
	findByEmailFn func(ctx context.Context, email string) (User, error)
}

func (m *mockRepository) Create(ctx context.Context, u User) (User, error) {
	return m.createFn(ctx, u)
}

func (m *mockRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	return m.findByEmailFn(ctx, email)
}

func TestService_Register_Success(t *testing.T) {
	var captured User
	repo := &mockRepository{
		findByEmailFn: func(ctx context.Context, email string) (User, error) {
			return User{}, errors.New("no rows")
		},
		createFn: func(ctx context.Context, u User) (User, error) {
			captured = u
			u.ID = 1
			u.CreatedAt = time.Now()
			return u, nil
		},
	}
	svc := NewService(repo, "test-secret")

	resp, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "user@example.com",
		Password: "correcthorse",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)

	claims := &jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(resp.AccessToken, claims, func(t *jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	require.NoError(t, err)
	assert.Equal(t, "1", claims.Subject)

	assert.Equal(t, "user@example.com", captured.Email)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(captured.PasswordHash), []byte("correcthorse")))
}

func TestService_Register_RepoError(t *testing.T) {
	repoErr := errors.New("insert failed")
	repo := &mockRepository{
		findByEmailFn: func(ctx context.Context, email string) (User, error) {
			return User{}, errors.New("no rows")
		},
		createFn: func(ctx context.Context, u User) (User, error) {
			return User{}, repoErr
		},
	}
	svc := NewService(repo, "test-secret")

	_, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "user@example.com",
		Password: "correcthorse",
	})

	assert.ErrorIs(t, err, repoErr)
}

func TestService_Register_EmailAlreadyExists(t *testing.T) {
	repo := &mockRepository{
		findByEmailFn: func(ctx context.Context, email string) (User, error) {
			return User{ID: 1, Email: email}, nil
		},
		createFn: func(ctx context.Context, u User) (User, error) {
			t.Fatal("Create should not be called when the email already exists")
			return User{}, nil
		},
	}
	svc := NewService(repo, "test-secret")

	_, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "user@example.com",
		Password: "correcthorse",
	})

	assert.Error(t, err)
}

func TestService_Login_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correcthorse"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo := &mockRepository{
		findByEmailFn: func(ctx context.Context, email string) (User, error) {
			return User{ID: 1, Email: email, PasswordHash: string(hash)}, nil
		},
	}
	svc := NewService(repo, "test-secret")

	resp, err := svc.Login(context.Background(), LoginRequest{
		Email:    "user@example.com",
		Password: "correcthorse",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestService_Login_WrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correcthorse"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo := &mockRepository{
		findByEmailFn: func(ctx context.Context, email string) (User, error) {
			return User{ID: 1, Email: email, PasswordHash: string(hash)}, nil
		},
	}
	svc := NewService(repo, "test-secret")

	_, err = svc.Login(context.Background(), LoginRequest{
		Email:    "user@example.com",
		Password: "wrongpassword",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestService_Login_UserNotFound(t *testing.T) {
	repo := &mockRepository{
		findByEmailFn: func(ctx context.Context, email string) (User, error) {
			return User{}, errors.New("no rows")
		},
	}
	svc := NewService(repo, "test-secret")

	_, err := svc.Login(context.Background(), LoginRequest{
		Email:    "ghost@example.com",
		Password: "whatever",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}
