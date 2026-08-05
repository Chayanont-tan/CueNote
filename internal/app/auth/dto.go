package auth

// RegisterRequest is the payload for creating a new account.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the payload for authenticating an existing account.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse carries the issued access token after register/login.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}
