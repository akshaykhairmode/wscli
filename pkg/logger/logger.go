package logger

import (
	"io"
	"strings"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/rs/zerolog"
)

var globalLogger *zerolog.Logger

var defaultFormatLevelFunc = func(i any) string {
	level := strings.ToUpper(i.(string))
	switch level {
	case "ERROR":
		return "[red]" + level + "[white]"
	case "DEBUG":
		return "[gray]" + level + "[white]"
	case "INFO":
		return "[blue]" + level + "[white]"
	}
	return level
}

func Init(out io.Writer, formatLevelFunc func(i any) string) {

	if formatLevelFunc == nil {
		formatLevelFunc = defaultFormatLevelFunc
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:         out,
		NoColor:     true,
		FormatLevel: formatLevelFunc,
		FormatTimestamp: func(i any) string {
			return time.Now().Format("15:04:05.000")
		},
	}

	l := zerolog.New(consoleWriter).With().Logger()

	if config.Flags.IsVerbose() {
		l = l.Level(zerolog.DebugLevel)
	} else {
		l = l.Level(zerolog.InfoLevel)
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
