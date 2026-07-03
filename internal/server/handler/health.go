package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Health struct {
	checks []func(ctx context.Context) error
}

func NewHealth(checks ...func(ctx context.Context) error) *Health {
	return &Health{checks: checks}
}

func (h *Health) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Health) Readiness(c echo.Context) error {
	ctx := c.Request().Context()

	for _, check := range h.checks {
		if err := check(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "not ready",
				"error":  err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}
