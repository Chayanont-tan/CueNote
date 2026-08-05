package auth

import "context"

// Repository defines persistence operations for the auth feature.
type Repository interface {
	Create(ctx context.Context, u User) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
}

// Service defines the business logic operations for the auth feature.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (TokenResponse, error)
	Login(ctx context.Context, req LoginRequest) (TokenResponse, error)
}
