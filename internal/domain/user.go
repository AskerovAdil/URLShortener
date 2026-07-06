package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if len(email) < 5 || !strings.Contains(email, "@") || strings.HasPrefix(email, "@") {
		return fmt.Errorf("%w: bad email", ErrInvalidInput)
	}
	return nil
}

func NewUser(email, passwordHash string) (*User, error) {
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}
	if passwordHash == "" {
		return nil, fmt.Errorf("%w: empty password hash", ErrInvalidInput)
	}

	return &User{
		ID:           uuid.New(),
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}, nil
}
