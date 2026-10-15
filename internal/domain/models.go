package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email or username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken      = errors.New("invalid or expired token")
	ErrTokenExpired      = errors.New("reset token has expired")
	ErrTokenAlreadyUsed  = errors.New("reset token has already been used")
	ErrSessionInvalidated = errors.New("session has been invalidated, please log in again")
)

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Credits      int64      `json:"credits"`
	Rating       int        `json:"rating"`
	TotalMatches int        `json:"total_matches"`
	Wins         int        `json:"wins"`
	TokenVersion int        `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type PasswordResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Username string `json:"username" binding:"required,min=3,max=30" example:"card_master"`
	Password string `json:"password" binding:"required,min=6,max=72" example:"secret123"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"card_master"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

type AuthResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  *User  `json:"user"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=72" example:"newsecret123"`
}

type MessageResponse struct {
	Message string `json:"message" example:"Операция выполнена успешно"`
}

type UserProfileResponse struct {
	ID           string `json:"id" example:"a4d2e482-1d54-47ef-a0b2-c07a9b8efaa1"`
	Email        string `json:"email" example:"user@example.com"`
	Username     string `json:"username" example:"card_master"`
	Credits      int64  `json:"credits" example:"5000"`
	Rating       int    `json:"rating" example:"1200"`
	TotalMatches int    `json:"total_matches" example:"0"`
	Wins         int    `json:"wins" example:"0"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	UpdatePassword(ctx context.Context, userID, newHash string) error
	IncrementTokenVersion(ctx context.Context, userID string) error
}

type PasswordResetRepository interface {
	CreateToken(ctx context.Context, resetToken *PasswordResetToken) error
	GetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkTokenAsUsed(ctx context.Context, tokenID string) error
}

type ResetLinkSender interface {
	SendPasswordResetLink(ctx context.Context, email, token string) error
}

type AuthService interface {
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	GetProfile(ctx context.Context, userID string) (*UserProfileResponse, error)
	ValidateToken(tokenString string) (userID string, tokenVersion int, err error)
}
