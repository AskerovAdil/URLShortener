package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Log        LogConfig        `mapstructure:"log"`
	App        AppConfig        `mapstructure:"app"`
	Postgres   PostgresConfig   `mapstructure:"postgres"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Cache      CacheConfig      `mapstructure:"cache"`
	Migrations MigrationsConfig `mapstructure:"migrations"`
	JWT        JWTConfig        `mapstructure:"jwt"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type LogConfig struct {
	Level       string `mapstructure:"level"`
	Development bool   `mapstructure:"development"`
}

type AppConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode,
	)
}

// MigrateDSN — golang-migrate pgx/v5 driver expects pgx5:// scheme.
func (p PostgresConfig) MigrateDSN() string {
	return fmt.Sprintf(
		"pgx5://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode,
	)
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type CacheConfig struct {
	URLTTL time.Duration `mapstructure:"url_ttl"`
}

type MigrationsConfig struct {
	Path         string `mapstructure:"path"`
	RunOnStartup bool   `mapstructure:"run_on_startup"`
}

type JWTConfig struct {
	Secret string        `mapstructure:"secret"`
	TTL    time.Duration `mapstructure:"ttl"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 10*time.Second)
	v.SetDefault("server.shutdown_timeout", 15*time.Second)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.development", false)
	v.SetDefault("app.base_url", "http://localhost:8080")
	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.user", "urlshortener")
	v.SetDefault("postgres.password", "urlshortener")
	v.SetDefault("postgres.database", "urlshortener")
	v.SetDefault("postgres.ssl_mode", "disable")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("cache.url_ttl", time.Hour)
	v.SetDefault("migrations.path", "migrations")
	v.SetDefault("migrations.run_on_startup", true)
	v.SetDefault("jwt.secret", "dev-secret-change-me-in-production-32chars")
	v.SetDefault("jwt.ttl", 24*time.Hour)

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config %q: %w", path, err)
		}
	}

	v.SetEnvPrefix("URLSHORTENER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
