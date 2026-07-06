package domain

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	AliasMinLen = 3
	AliasMaxLen = 32
)

var aliasPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type URL struct {
	ID          uuid.UUID
	Alias       string
	OriginalURL string
	UserID      uuid.UUID
	ExpiresAt   *time.Time
	CreatedAt   time.Time
}

func ValidateAlias(alias string) error {
	if len(alias) < AliasMinLen || len(alias) > AliasMaxLen {
		return fmt.Errorf("%w: alias length must be %d..%d", ErrInvalidInput, AliasMinLen, AliasMaxLen)
	}
	if !aliasPattern.MatchString(alias) {
		return fmt.Errorf("%w: alias has invalid characters", ErrInvalidInput)
	}
	return nil
}

func ValidateOriginalURL(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("%w: malformed url", ErrInvalidInput)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: only http/https allowed", ErrInvalidInput)
	}
	if u.Host == "" {
		return fmt.Errorf("%w: url host is required", ErrInvalidInput)
	}
	return nil
}

func NewURL(alias, originalURL string, userID uuid.UUID, expiresAt *time.Time) (*URL, error) {
	if err := ValidateAlias(alias); err != nil {
		return nil, err
	}
	if err := ValidateOriginalURL(originalURL); err != nil {
		return nil, err
	}
	if userID == uuid.Nil {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	return &URL{
		ID:          uuid.New(),
		Alias:       alias,
		OriginalURL: originalURL,
		UserID:      userID,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (u *URL) IsExpired(now time.Time) bool {
	return u.ExpiresAt != nil && !u.ExpiresAt.After(now)
}
