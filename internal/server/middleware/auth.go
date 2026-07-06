package middleware

import (
	"net/http"
	"strings"

	"github.com/AskerovAdil/URLShortener/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const userIDKey = "user_id"

func JWT(auth *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get(echo.HeaderAuthorization)
			if header == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
			}

			userID, err := auth.ParseToken(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			c.Set(userIDKey, userID)
			return next(c)
		}
	}
}

func UserID(c echo.Context) (uuid.UUID, bool) {
	v, ok := c.Get(userIDKey).(uuid.UUID)
	return v, ok && v != uuid.Nil
}
