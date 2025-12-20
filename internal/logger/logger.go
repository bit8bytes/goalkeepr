package logger

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bit8bytes/goalkeepr/pkg/slogtrace"
)

// Setup creates a slog.Logger with simple slogtrace integration
func Setup(level *slog.LevelVar) *slog.Logger {
	loggerOpts := &slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: includeSourceFile,
	}

	handler := slogtrace.NewJSONHandler(os.Stdout, loggerOpts)
	return slog.New(handler)
}

// This is not helpful when calling a helper function
// but it helps in general.
func includeSourceFile(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.File = filepath.Base(source.File)
	}
	return a
}
