package logging

import (
	"os"

	"github.com/rs/zerolog"
)

func New() zerolog.Logger {
	return zerolog.New(os.Stdout).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Caller().
		Logger()
}
