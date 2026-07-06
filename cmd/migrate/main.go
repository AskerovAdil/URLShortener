package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AskerovAdil/URLShortener/internal/config"
	"github.com/AskerovAdil/URLShortener/internal/migrate"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch *direction {
	case "up":
		err = migrate.Up(cfg.Postgres.MigrateDSN(), cfg.Migrations.Path)
	case "down":
		err = migrate.Down(cfg.Postgres.MigrateDSN(), cfg.Migrations.Path)
	default:
		_, _ = fmt.Fprintln(os.Stderr, "direction must be up or down")
		os.Exit(1)
	}

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
