package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"durak-online/internal/domain"
	"durak-online/pkg/hasher"
	"durak-online/pkg/token"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo        domain.UserRepository
	resetRepo       domain.PasswordResetRepository
	hasher          hasher.PasswordHasher
	jwtManager      *token.JWTManager
	resetLinkSender domain.ResetLinkSender
	resetTokenTTL   time.Duration
}

func NewAuthService(
	userRepo domain.UserRepository,
	resetRepo domain.PasswordResetRepository,
	hasher hasher.PasswordHasher,
	jwtManager *token.JWTManager,
	resetLinkSender domain.ResetLinkSender,
	resetTokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		resetRepo:       resetRepo,
		hasher:          hasher,
		jwtManager:      jwtManager,
		resetLinkSender: resetLinkSender,
		resetTokenTTL:   resetTokenTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, domain.ErrUserAlreadyExists
	}
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hash,
		Credits:      5000,
		Rating:       1200,
		TotalMatches: 0,
		Wins:         0,
		TokenVersion: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	tokenStr, err := s.jwtManager.Generate(user.ID, user.TokenVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	return &domain.AuthResponse{
		Token: tokenStr,
		User:  user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	req.Username = strings.TrimSpace(req.Username)

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := s.hasher.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokenStr, err := s.jwtManager.Generate(user.ID, user.TokenVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	return &domain.AuthResponse{
		Token: tokenStr,
		User:  user,
	}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest) error {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil
		}
		return err
	}

	resetToken := &domain.PasswordResetToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(s.resetTokenTTL),
		CreatedAt: time.Now(),
	}

	if err := s.resetRepo.CreateToken(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	_ = s.resetLinkSender.SendPasswordResetLink(ctx, user.Email, resetToken.Token)

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error {
	tokenRecord, err := s.resetRepo.GetToken(ctx, req.Token)
	if err != nil {
		return domain.ErrInvalidToken
	}

	if tokenRecord.UsedAt != nil {
		return domain.ErrTokenAlreadyUsed
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		return domain.ErrTokenExpired
	}

	newHash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, tokenRecord.UserID, newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := s.userRepo.IncrementTokenVersion(ctx, tokenRecord.UserID); err != nil {
		return fmt.Errorf("failed to invalidate active sessions: %w", err)
	}

	if err := s.resetRepo.MarkTokenAsUsed(ctx, tokenRecord.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID string) (*domain.UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.UserProfileResponse{
		ID:           user.ID,
		Email:        user.Email,
		Username:     user.Username,
		Credits:      user.Credits,
		Rating:       user.Rating,
		TotalMatches: user.TotalMatches,
		Wins:         user.Wins,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (string, int, error) {
	claims, err := s.jwtManager.Verify(tokenString)
	if err != nil {
		return "", 0, domain.ErrInvalidToken
	}
	return claims.UserID, claims.TokenVersion, nil
}
