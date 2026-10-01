package main

import (
	"github.com/NicoYazawa/Orbis/internal/phase0"
	"log/slog"
	"os"
)

func main() {
	if err := phase0.RunHTTP("api", ":8080"); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
