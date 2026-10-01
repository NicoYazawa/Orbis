package main

import (
	"github.com/NicoYazawa/Orbis/internal/phase0"
	"log/slog"
	"os"
)

func main() {
	if err := phase0.RunHTTP("worker", ":8081"); err != nil {
		slog.Error("worker failed", "error", err)
		os.Exit(1)
	}
}
