package resettoken

import (
	"context"
	"database/sql"
	"errors"

	"durak-online/internal/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateToken(ctx context.Context, t *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.UserID, t.Token, t.ExpiresAt, t.CreatedAt)
	return err
}

func (r *Repository) GetToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`
	var t domain.PasswordResetToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}
	return &t, nil
}

func (r *Repository) MarkTokenAsUsed(ctx context.Context, tokenID string) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, tokenID)
	return err
}
