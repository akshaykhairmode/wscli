package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/akshaykhairmode/wscli/pkg/batch"
	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/akshaykhairmode/wscli/pkg/ws"
	"github.com/chzyer/readline"
	"github.com/gorilla/websocket"
)

func main() {

	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	cfg := config.Get()

	logger.Init(cfg)

	readlineConfig := &readline.Config{
		Prompt:          batch.GetPrompt(cfg, "» "),
		AutoComplete:    completer,
		HistoryFile:     getHistoryFilePath("wscli"),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
	}

	rl, err := readline.NewEx(readlineConfig)
	if err != nil {
		logger.GlobalLogger.Fatal().Err(err).Msg("error while creating readline object")
	}

	defer rl.Close()

	conn, closeFunc, err := ws.Connect(cfg)
	if err != nil {
		logger.GlobalLogger.Fatal().Err(err).Msg("CONNECT ERR")
	}

	defer closeFunc()

	if !cfg.Stdin {
		rl.CaptureExitSignal()
	} else {
		go catchSignals(conn)
	}

	batch.Process(conn, cfg, rl)
}

func catchSignals(conn *websocket.Conn) {
	sigs := make(chan os.Signal, 2)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	logger.GlobalLogger.Debug().Msgf("received signal %s", <-sigs)

	if err := conn.Close(); err != nil {
		logger.GlobalLogger.Debug().Err(err).Msg("error while closing connection")
	}
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("/connect"),
	readline.PcItem("/exit"),
	readline.PcItem("/ping"),
	readline.PcItem("/pong"),
	readline.PcItem("/wait"),
	readline.PcItem("/help"),
	readline.PcItem("/print",
		readline.PcItem("host"),
	),
)

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

func getHistoryFilePath(appName string) string {

	fallback := ".readline.history"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.GlobalLogger.Debug().Err(err).Msg("error while getting homedir")
		return fallback
	}

	var historyPath string
	switch runtime.GOOS {
	case "linux", "darwin":
		configDir := filepath.Join(homeDir, ".config", appName)
		err := os.MkdirAll(configDir, 0700)
		if err != nil {
			logger.GlobalLogger.Debug().Err(err).Msg("error while creating directory in linux,darwin os")
			return fallback
		}
		historyPath = filepath.Join(configDir, "history")
	case "windows":
		appData := os.Getenv("AppData")
		configDir := filepath.Join(appData, appName)
		err := os.MkdirAll(configDir, 0700)
		if err != nil {
			logger.GlobalLogger.Debug().Err(err).Msg("error while creating directory in windows os")
			return fallback
		}
		historyPath = filepath.Join(configDir, "history")
	default:
		historyPath = filepath.Join(homeDir, "."+appName+"_history")
	}

	logger.GlobalLogger.Debug().Msgf("History Path is %s", historyPath)

	return historyPath
}
