package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Health struct{}

func NewHealth() *Health {
	return &Health{}
}

func (h *Health) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Health) Readiness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}