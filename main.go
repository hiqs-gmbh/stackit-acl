package main

import (
	"log/slog"
	"os"

	"stackit-acl/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
