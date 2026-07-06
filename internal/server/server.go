package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/AskerovAdil/URLShortener/internal/config"
	"github.com/AskerovAdil/URLShortener/internal/server/handler"
	authmw "github.com/AskerovAdil/URLShortener/internal/server/middleware"
	"github.com/AskerovAdil/URLShortener/internal/service"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Deps struct {
	Auth             *handler.Auth
	AuthService      *service.AuthService
	ReadinessChecks  []func(ctx context.Context) error
}

func New(cfg *config.Config, log *zap.Logger, deps Deps) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.HTTPErrorHandler = errorHandler(log)

	e.Use(echomw.Recover())
	e.Use(requestLogger(log))
	e.Use(echomw.RequestID())

	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout

	health := handler.NewHealth(deps.ReadinessChecks...)

	e.GET("/health", health.Liveness)
	e.GET("/ready", health.Readiness)

	v1 := e.Group("/api/v1")

	authGroup := v1.Group("/auth")
	authGroup.POST("/register", deps.Auth.Register)
	authGroup.POST("/login", deps.Auth.Login)

	protected := v1.Group("")
	protected.Use(authmw.JWT(deps.AuthService))
	protected.GET("/me", deps.Auth.Me)

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

		if err := c.JSON(code, map[string]string{"error": msg}); err != nil {
			log.Warn("failed to write error response", zap.Error(err))
		}
	}
}
