package main

import (
	"log/slog"
	"os"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	log.Info("Hello World!")
}