package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"log/slog"
)

var (
	colorStep    = color.New(color.Faint)
	colorInfo    = color.New(color.FgCyan)
	colorSuccess = color.New(color.FgGreen)
	colorWarn    = color.New(color.FgYellow)
	colorError   = color.New(color.FgRed, color.Bold)
	colorPrompt  = color.New(color.Bold)
	colorAdded   = color.New(color.FgGreen)
	colorRemoved = color.New(color.FgRed)
)

type cliHandler struct {
	level slog.Level
	mu    *sync.Mutex
}

func newCLIHandler(level slog.Level) *cliHandler {
	return &cliHandler{level: level, mu: &sync.Mutex{}}
}

func (h *cliHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *cliHandler) Handle(_ context.Context, r slog.Record) error {
	kind := "info"
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "kind" {
			kind = a.Value.String()
		}
		return true
	})

	var prefix string
	var c *color.Color

	switch kind {
	case "step":
		prefix = "→ "
		c = colorStep
	case "success":
		prefix = "✓ "
		c = colorSuccess
	case "warn":
		prefix = "⚠ "
		c = colorWarn
	case "error":
		prefix = "✗ "
		c = colorError
	default:
		c = colorInfo
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if r.Level >= slog.LevelError {
		_, _ = c.Fprintf(os.Stderr, "%s%s\n", prefix, r.Message)
	} else {
		_, _ = c.Fprintf(os.Stdout, "%s%s\n", prefix, r.Message)
	}

	return nil
}

func (h *cliHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *cliHandler) WithGroup(_ string) slog.Handler {
	return h
}

func setupLogger(v string) {
	level := slog.LevelInfo
	switch v {
	case "error":
		level = slog.LevelError
	case "warning":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	case "debug":
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(newCLIHandler(level)))
}

func logStep(msg string, args ...any) {
	slog.Info(msg, append([]any{"kind", "step"}, args...)...)
}

func logInfo(msg string, args ...any) {
	slog.Info(msg, args...)
}

func logSuccess(msg string, args ...any) {
	slog.Info(msg, append([]any{"kind", "success"}, args...)...)
}

func logWarn(msg string, args ...any) {
	slog.Warn(msg, append([]any{"kind", "warn"}, args...)...)
}

func outPrompt(msg string) {
	_, _ = colorPrompt.Print(msg + " ")
}

func formatACLList(current []string, changed string, act action) string {
	var sb strings.Builder
	for _, entry := range current {
		if entry == changed {
			if act == actionAdd {
				sb.WriteString("  ")
				_, _ = colorAdded.Fprint(&sb, "+ "+entry)
				sb.WriteString("\n")
			} else {
				sb.WriteString("  ")
				_, _ = colorRemoved.Fprint(&sb, "- "+entry)
				sb.WriteString("\n")
			}
		} else {
			fmt.Fprintf(&sb, "  • %s\n", entry)
		}
	}
	return sb.String()
}
