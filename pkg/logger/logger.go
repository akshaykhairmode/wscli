package logger

import (
	"os"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/rs/zerolog"
)

var GlobalLogger *zerolog.Logger

func Init(cfg config.Config) {
	consoleWriter := zerolog.ConsoleWriter{
		Out:             os.Stderr,
		NoColor:         cfg.NoColor,
		FormatTimestamp: func(i interface{}) string { return "" },
		// FormatMessage:   func(i interface{}) string { return "" },
	}

	var l zerolog.Logger

	if cfg.Verbose {
		l = zerolog.New(consoleWriter).With().Logger().Level(zerolog.DebugLevel)
	} else {
		l = zerolog.New(consoleWriter).With().Logger().Level(zerolog.InfoLevel)
	}

	GlobalLogger = &l

}
