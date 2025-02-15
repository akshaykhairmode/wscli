package logger

import (
	"os"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/rs/zerolog"
)

var GlobalLogger *zerolog.Logger

func Init(cfg config.Config) {
	// Create a console writer with human-readable format
	consoleWriter := zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: cfg.NoColor,
	}

	// Create a logger with the console writer
	l := zerolog.New(consoleWriter).With().Logger()
	GlobalLogger = &l
}
