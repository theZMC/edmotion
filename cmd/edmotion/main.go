package main

import (
	"log/slog"
	"os"

	"github.com/thezmc/edmotion/internal/config"
	"github.com/thezmc/edmotion/internal/runtime"
)

func main() {
	settings := config.Load()
	app := runtime.New(settings)

	if err := app.Run(); err != nil {
		slog.Error("application error", slog.Any("error", err))
		os.Exit(1)
	}
}
