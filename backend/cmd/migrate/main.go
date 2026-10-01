package main

import (
	"context"
	"github.com/NicoYazawa/Orbis/internal/phase0"
	"log/slog"
	"os"
	"time"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	c, err := phase0.LoadConfig(os.Getenv)
	if err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	d, err := phase0.Open(ctx, c)
	if err != nil {
		slog.Error("dependency connection failed")
		os.Exit(1)
	}
	defer d.Close()
	if err := d.Initialize(ctx); err != nil {
		slog.Error("initialization failed")
		os.Exit(1)
	}
	slog.Info("initialization complete")
}
