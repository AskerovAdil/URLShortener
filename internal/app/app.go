package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AskerovAdil/URLShortener/internal/cache/redis"
	"github.com/AskerovAdil/URLShortener/internal/config"
	"github.com/AskerovAdil/URLShortener/internal/migrate"
	"github.com/AskerovAdil/URLShortener/internal/repository/postgres"
	"github.com/AskerovAdil/URLShortener/internal/server"
	"github.com/AskerovAdil/URLShortener/internal/server/handler"
	"github.com/AskerovAdil/URLShortener/internal/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	log, err := newLogger(cfg.Log)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	ctx := context.Background()

	pg, err := postgres.Open(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer pg.Close()

	if cfg.Migrations.RunOnStartup {
		log.Info("running migrations", zap.String("path", cfg.Migrations.Path))
		if err := migrate.Up(cfg.Postgres.MigrateDSN(), cfg.Migrations.Path); err != nil {
			return fmt.Errorf("migrations: %w", err)
		}
	}

	rdb, err := redis.Open(ctx, cfg.Redis)
	if err != nil {
		return err
	}
	defer func() { _ = rdb.Close() }()

	userRepo := postgres.NewUserRepo(pg)

	authSvc, err := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.TTL)
	if err != nil {
		return fmt.Errorf("init auth: %w", err)
	}

	authHandler := handler.NewAuth(authSvc)

	urlRepo := postgres.NewURLRepo(pg)
	urlCache := redis.NewURLCache(rdb, cfg.Cache.URLTTL)
	urlSvc := service.NewURLService(urlRepo, urlCache, cfg.Cache.URLTTL)
	urlHandler := handler.NewURL(urlSvc, cfg.App.BaseURL)

	e := server.New(cfg, log, server.Deps{
		Auth:        authHandler,
		URL:         urlHandler,
		AuthService: authSvc,
		ReadinessChecks: []func(context.Context) error{
			func(c context.Context) error { return pg.Ping(c) },
			func(c context.Context) error { return rdb.Ping(c).Err() },
		},
	})

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info("server starting", zap.String("addr", cfg.Server.Addr()))
		if err := e.Start(cfg.Server.Addr()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	case <-runCtx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	log.Info("server stopped")
	return nil
}

func newLogger(cfg config.LogConfig) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", cfg.Level, err)
	}

	zcfg := zap.Config{
		Level:            level,
		Development:      cfg.Development,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if cfg.Development {
		zcfg.Encoding = "console"
		zcfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return zcfg.Build()
}
