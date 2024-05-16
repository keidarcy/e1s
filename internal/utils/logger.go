package utils

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

// GetLogger returns a *slog.Logger configured to write to the specified file path.
// It also returns the log file *os.File  itself, such that callers can close the
// file if/when needed.
func GetLogger(path string, json bool, debug bool) (*slog.Logger, *os.File) {
	option := slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}
	if debug {
		option.Level = slog.LevelDebug
	}

	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		out = os.Stdout
		fmt.Printf("failed to log to file %s, using default stderr, err: %v", path, err)
	}
	var handler slog.Handler
	if json {
		handler = slog.NewJSONHandler(out, &option)
	} else {
		handler = tint.NewHandler(out, &tint.Options{
			AddSource:  option.AddSource,
			Level:      option.Level,
			TimeFormat: time.RFC3339,
		})
	}

	logger := slog.New(handler)
	return logger, out
}
