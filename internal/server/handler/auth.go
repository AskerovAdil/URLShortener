package handler

import (
	"errors"
	"net/http"

	"github.com/AskerovAdil/URLShortener/internal/domain"
	"github.com/AskerovAdil/URLShortener/internal/server/middleware"
	"github.com/AskerovAdil/URLShortener/internal/service"
	"github.com/labstack/echo/v4"
)

type Auth struct {
	svc *service.AuthService
}

func NewAuth(svc *service.AuthService) *Auth {
	return &Auth{svc: svc}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

func (h *Auth) Register(c echo.Context) error {
	var req authRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json body")
	}

	token, err := h.svc.Register(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return mapAuthError(err)
	}

	return c.JSON(http.StatusCreated, tokenResponse{Token: token})
}

func (h *Auth) Login(c echo.Context) error {
	var req authRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json body")
	}

	token, err := h.svc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return mapAuthError(err)
	}

	return c.JSON(http.StatusOK, tokenResponse{Token: token})
}

// Me — sanity-check для JWT, пригодится пока нет URL endpoints.
func (h *Auth) Me(c echo.Context) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"user_id": userID.String(),
	})
}

func mapAuthError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrConflict):
		return echo.NewHTTPError(http.StatusConflict, "email already registered")
	case errors.Is(err, domain.ErrUnauthorized):
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
