package main

import (
	"flag"
	"os"

	"github.com/AskerovAdil/URLShortener/internal/app"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	if err := app.Run(*configPath); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
