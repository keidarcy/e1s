package view

import (
	"log/slog"
	"testing"
)

func TestMain(m *testing.M) {
	logger = slog.New()
}
