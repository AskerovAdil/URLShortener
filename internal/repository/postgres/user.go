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

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, q, user.ID, user.Email, user.PasswordHash, user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "users_email_unique" {
			return domain.ErrConflict
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var u domain.User
	err := r.pool.QueryRow(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	var u domain.User
	err := r.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}

	return &u, nil
}
