package main

import (
	"log/slog"
	"os"

	"github.com/rincewind_log/internal/command"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	command.Execute()
}
