package helpers

import (
	"os"

	"github.com/rs/zerolog"
)

// Log is the global logger
var Log zerolog.Logger

// SetupLogger configures the global logger
func SetupLogger(level string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	Log = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"},
	).With().Timestamp().Logger().Level(lvl)
}
