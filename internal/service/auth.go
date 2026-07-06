package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AskerovAdil/URLShortener/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type AuthService struct {
	users  UserStore
	secret []byte
	ttl    time.Duration
}

func NewAuthService(users UserStore, secret string, ttl time.Duration) (*AuthService, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("jwt secret must be at least 32 characters")
	}

	return &AuthService{
		users:  users,
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

type tokenClaims struct {
	UserID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {
	if err := validatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user, err := domain.NewUser(email, string(hash))
	if err != nil {
		return "", err
	}

	if err := s.users.Create(ctx, user); err != nil {
		return "", err
	}

	return s.issueToken(user.ID)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	if err := domain.ValidateEmail(email); err != nil {
		return "", domain.ErrUnauthorized
	}

	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return "", domain.ErrUnauthorized
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrUnauthorized
	}

	return s.issueToken(user.ID)
}

func (s *AuthService) ParseToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &tokenClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return uuid.Nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid || claims.UserID == uuid.Nil {
		return uuid.Nil, domain.ErrUnauthorized
	}

	return claims.UserID, nil
}

func (s *AuthService) issueToken(userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := tokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", domain.ErrInvalidInput)
	}
	return nil
}
