package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AskerovAdil/URLShortener/internal/domain"
	"github.com/AskerovAdil/URLShortener/internal/server/middleware"
	"github.com/AskerovAdil/URLShortener/internal/service"
	"github.com/labstack/echo/v4"
)

type URL struct {
	svc     *service.URLService
	baseURL string
}

func NewURL(svc *service.URLService, baseURL string) *URL {
	return &URL{
		svc:     svc,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

type createURLRequest struct {
	OriginalURL string     `json:"original_url"`
	Alias       string     `json:"alias"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type urlResponse struct {
	Alias       string     `json:"alias"`
	OriginalURL string     `json:"original_url"`
	ShortURL    string     `json:"short_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (h *URL) Create(c echo.Context) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req createURLRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json body")
	}

	u, err := h.svc.Create(c.Request().Context(), userID, service.CreateURLInput{
		OriginalURL: req.OriginalURL,
		Alias:       req.Alias,
		ExpiresAt:   req.ExpiresAt,
	})
	if err != nil {
		return mapURLError(err)
	}

	return c.JSON(http.StatusCreated, toURLResponse(h.baseURL, u))
}

func (h *URL) List(c echo.Context) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	urls, err := h.svc.List(c.Request().Context(), userID)
	if err != nil {
		return mapURLError(err)
	}

	resp := make([]urlResponse, 0, len(urls))
	for _, u := range urls {
		resp = append(resp, toURLResponse(h.baseURL, u))
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *URL) Delete(c echo.Context) error {
	userID, ok := middleware.UserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	alias := c.Param("alias")
	if err := h.svc.Delete(c.Request().Context(), userID, alias); err != nil {
		return mapURLError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *URL) Redirect(c echo.Context) error {
	alias := c.Param("alias")

	original, err := h.svc.Resolve(c.Request().Context(), alias)
	if err != nil {
		return mapURLError(err)
	}

	return c.Redirect(http.StatusFound, original)
}

func toURLResponse(baseURL string, u *domain.URL) urlResponse {
	return urlResponse{
		Alias:       u.Alias,
		OriginalURL: u.OriginalURL,
		ShortURL:    baseURL + "/" + u.Alias,
		ExpiresAt:   u.ExpiresAt,
		CreatedAt:   u.CreatedAt,
	}
}

func mapURLError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrConflict):
		return echo.NewHTTPError(http.StatusConflict, "alias already taken")
	case errors.Is(err, domain.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "link not found")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
