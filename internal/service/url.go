package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AskerovAdil/URLShortener/internal/domain"
	"github.com/AskerovAdil/URLShortener/internal/pkg/shortcode"
	"github.com/google/uuid"
)

const maxAliasAttempts = 5

type URLRepository interface {
	Create(ctx context.Context, u *domain.URL) error
	GetByAlias(ctx context.Context, alias string) (*domain.URL, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.URL, error)
	Delete(ctx context.Context, alias string, userID uuid.UUID) error
}

type URLCacheStore interface {
	Get(ctx context.Context, alias string) (string, error)
	Set(ctx context.Context, alias, originalURL string, ttl time.Duration) error
	Delete(ctx context.Context, alias string) error
}

type URLService struct {
	repo     URLRepository
	cache    URLCacheStore
	cacheTTL time.Duration
}

func NewURLService(repo URLRepository, cache URLCacheStore, cacheTTL time.Duration) *URLService {
	return &URLService{
		repo:     repo,
		cache:    cache,
		cacheTTL: cacheTTL,
	}
}

type CreateURLInput struct {
	OriginalURL string
	Alias       string
	ExpiresAt   *time.Time
}

func (s *URLService) Create(ctx context.Context, userID uuid.UUID, in CreateURLInput) (*domain.URL, error) {
	alias := strings.TrimSpace(in.Alias)

	if alias == "" {
		return s.createWithGeneratedAlias(ctx, userID, in.OriginalURL, in.ExpiresAt)
	}

	u, err := domain.NewURL(alias, in.OriginalURL, userID, in.ExpiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *URLService) createWithGeneratedAlias(
	ctx context.Context,
	userID uuid.UUID,
	originalURL string,
	expiresAt *time.Time,
) (*domain.URL, error) {
	var lastErr error

	for range maxAliasAttempts {
		alias, err := shortcode.Generate(8)
		if err != nil {
			return nil, err
		}

		u, err := domain.NewURL(alias, originalURL, userID, expiresAt)
		if err != nil {
			return nil, err
		}

		if err := s.repo.Create(ctx, u); err != nil {
			if errors.Is(err, domain.ErrConflict) {
				lastErr = err
				continue
			}
			return nil, err
		}

		return u, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("failed to generate unique alias")
}

func (s *URLService) Resolve(ctx context.Context, alias string) (string, error) {
	if original, err := s.cache.Get(ctx, alias); err == nil {
		return original, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return "", err
	}

	u, err := s.repo.GetByAlias(ctx, alias)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	if u.IsExpired(now) {
		return "", domain.ErrNotFound
	}

	ttl := s.cacheTTL
	if u.ExpiresAt != nil {
		if until := u.ExpiresAt.Sub(now); until < ttl {
			ttl = until
		}
	}

	// cache miss fill — best effort, не блокируем redirect
	if ttl > 0 {
		_ = s.cache.Set(ctx, alias, u.OriginalURL, ttl)
	}

	return u.OriginalURL, nil
}

func (s *URLService) List(ctx context.Context, userID uuid.UUID) ([]*domain.URL, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *URLService) Delete(ctx context.Context, userID uuid.UUID, alias string) error {
	if err := s.repo.Delete(ctx, alias, userID); err != nil {
		return err
	}

	_ = s.cache.Delete(ctx, alias)
	return nil
}
