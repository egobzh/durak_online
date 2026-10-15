package user

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

func (r *Repository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, credits, rating, total_matches, wins, token_version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		u.ID, u.Email, u.Username, u.PasswordHash, u.Credits, u.Rating,
		u.TotalMatches, u.Wins, u.TokenVersion, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, credits, rating, total_matches, wins, token_version, created_at, updated_at
		FROM users WHERE id = $1
	`
	return scanUser(r.db.QueryRowContext(ctx, query, id))
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, credits, rating, total_matches, wins, token_version, created_at, updated_at
		FROM users WHERE LOWER(email) = LOWER($1)
	`
	return scanUser(r.db.QueryRowContext(ctx, query, email))
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, credits, rating, total_matches, wins, token_version, created_at, updated_at
		FROM users WHERE LOWER(username) = LOWER($1)
	`
	return scanUser(r.db.QueryRowContext(ctx, query, username))
}

func (r *Repository) UpdatePassword(ctx context.Context, userID, newHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	res, err := r.db.ExecContext(ctx, query, newHash, userID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *Repository) IncrementTokenVersion(ctx context.Context, userID string) error {
	query := `UPDATE users SET token_version = token_version + 1, updated_at = NOW() WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash,
		&u.Credits, &u.Rating, &u.TotalMatches, &u.Wins,
		&u.TokenVersion, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}
