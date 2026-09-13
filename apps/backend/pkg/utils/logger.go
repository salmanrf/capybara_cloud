package utils

import (
	"log/slog"
	"os"
	"sync"
)

var Logger slog.Logger
var logsetup sync.Once

func CreateLogger() {
	logsetup.Do(func () {
		Logger = *slog.New(
			slog.NewTextHandler(
			os.Stderr,
			nil,
		))
	})
}