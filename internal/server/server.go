package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/AskerovAdil/URLShortener/internal/config"
	"github.com/AskerovAdil/URLShortener/internal/server/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func New(cfg *config.Config, log *zap.Logger) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.HTTPErrorHandler = errorHandler(log)

	e.Use(middleware.Recover())
	e.Use(requestLogger(log))
	e.Use(middleware.RequestID())

	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout

	health := handler.NewHealth()

	e.GET("/health", health.Liveness)
	e.GET("/ready", health.Readiness)

	return e
}

func requestLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			req := c.Request()
			res := c.Response()

			fields := []zap.Field{
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.Int("status", res.Status),
				zap.Duration("latency", time.Since(start)),
				zap.String("request_id", res.Header().Get(echo.HeaderXRequestID)),
			}
			if err != nil {
				fields = append(fields, zap.Error(err))
			}

			if res.Status >= 500 {
				log.Error("request", fields...)
			} else {
				log.Info("request", fields...)
			}

			return err
		}
	}
}

func errorHandler(log *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		code := http.StatusInternalServerError
		msg := "internal server error"

		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
			switch m := he.Message.(type) {
			case string:
				msg = m
			case error:
				msg = m.Error()
			default:
				msg = http.StatusText(code)
			}
		}

		if code >= 500 {
			log.Error("unhandled error",
				zap.Error(err),
				zap.String("path", c.Request().URL.Path),
			)
		}

		// don't write body if client already gone
		if err := c.JSON(code, map[string]string{"error": msg}); err != nil {
			log.Warn("failed to write error response", zap.Error(err))
		}
	}
}
