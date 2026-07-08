package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AskerovAdil/URLShortener/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRepo struct {
	pool *pgxpool.Pool
}

func NewURLRepo(pool *pgxpool.Pool) *URLRepo {
	return &URLRepo{pool: pool}
}

func (r *URLRepo) Create(ctx context.Context, u *domain.URL) error {
	const q = `
		INSERT INTO urls (id, alias, original_url, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, q, u.ID, u.Alias, u.OriginalURL, u.UserID, u.ExpiresAt, u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "urls_alias_unique" {
			return domain.ErrConflict
		}
		return fmt.Errorf("insert url: %w", err)
	}

	return nil
}

func (r *URLRepo) GetByAlias(ctx context.Context, alias string) (*domain.URL, error) {
	const q = `
		SELECT id, alias, original_url, user_id, expires_at, created_at
		FROM urls
		WHERE alias = $1
	`

	var u domain.URL
	err := r.pool.QueryRow(ctx, q, alias).Scan(
		&u.ID, &u.Alias, &u.OriginalURL, &u.UserID, &u.ExpiresAt, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select url: %w", err)
	}

	return &u, nil
}

func (r *URLRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.URL, error) {
	const q = `
		SELECT id, alias, original_url, user_id, expires_at, created_at
		FROM urls
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list urls: %w", err)
	}
	defer rows.Close()

	var urls []*domain.URL
	for rows.Next() {
		var u domain.URL
		if err := rows.Scan(&u.ID, &u.Alias, &u.OriginalURL, &u.UserID, &u.ExpiresAt, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan url: %w", err)
		}
		urls = append(urls, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate urls: %w", err)
	}

	return urls, nil
}

func (r *URLRepo) Delete(ctx context.Context, alias string, userID uuid.UUID) error {
	const q = `DELETE FROM urls WHERE alias = $1 AND user_id = $2`

	tag, err := r.pool.Exec(ctx, q, alias, userID)
	if err != nil {
		return fmt.Errorf("delete url: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
