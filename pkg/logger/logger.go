package logger

import (
	"os"
	"strings"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/rs/zerolog"
)

var globalLogger *zerolog.Logger

const logPrefix = "WSCLI:: "

func Init() {
	consoleWriter := zerolog.ConsoleWriter{
		Out:             os.Stderr,
		NoColor:         config.Flags.IsNoColor(),
		FormatTimestamp: func(i any) string { return "" },
		FormatMessage: func(i any) string {
			msg := i.(string)
			if !strings.HasPrefix(msg, logPrefix) {
				return logPrefix + msg
			}
			return msg
		},
		FormatLevel: func(i any) string {
			return ""
		},
	}

	var l zerolog.Logger

	if config.Flags.IsVerbose() {
		l = zerolog.New(consoleWriter).With().Logger().Level(zerolog.DebugLevel)
	} else {
		l = zerolog.New(consoleWriter).With().Logger().Level(zerolog.InfoLevel)
	}

	globalLogger = &l

}

func Debug() *zerolog.Event {
	return globalLogger.Debug()
}

func Fatal() *zerolog.Event {
	return globalLogger.Fatal()
}

func Error() *zerolog.Event {
	return globalLogger.Error()
}

func Info() *zerolog.Event {
	return globalLogger.Info()
}

func Err(err error) *zerolog.Event {
	return globalLogger.Err(err)
}
