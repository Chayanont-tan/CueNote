package auth

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
)

const tokenTTL = 7 * 24 * time.Hour

type service struct {
	repo      Repository
	jwtSecret string
}

// NewService creates the auth Service.
func NewService(repo Repository, jwtSecret string) Service {
	return &service{repo: repo, jwtSecret: jwtSecret}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (TokenResponse, error) {
	_, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil {
		return TokenResponse{}, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return TokenResponse{}, err
	}

	user, err := s.repo.Create(ctx, User{
		Email:        req.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return TokenResponse{}, err
	}

	return s.issueToken(user)
}

func (s *service) Login(ctx context.Context, req LoginRequest) (TokenResponse, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return TokenResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return TokenResponse{}, ErrInvalidCredentials
	}

	return s.issueToken(user)
}

// issueToken signs a JWT with the user's ID as the subject claim.
func (s *service) issueToken(user User) (TokenResponse, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(user.ID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{AccessToken: signed}, nil
}
