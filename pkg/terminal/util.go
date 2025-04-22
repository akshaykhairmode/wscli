package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/chzyer/readline"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("/connect"),
	readline.PcItem("/exit"),
	readline.PcItem("/ping"),
	readline.PcItem("/pong"),
	readline.PcItem("/wait"),
	readline.PcItem("/help"),
	readline.PcItem("/flags"),
	readline.PcItem("/print"),
)

func getDefaultConfig() *readline.Config {
	return &readline.Config{
		Prompt:          getPrompt("» "),
		AutoComplete:    completer,
		HistoryFile:     getHistoryFilePath("wscli"),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
	}
}

func getPrompt(str string) string {
	if config.Flags.NoColor {
		return str
	}

	return fmt.Sprintf("\033[31m%s\033[0m ", str)
}

func getHistoryFilePath(appName string) string {

	fallback := ".readline.history"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Debug().Err(err).Msg("error while getting homedir")
		return fallback
	}

	var historyPath string
	switch runtime.GOOS {
	case "linux", "darwin":
		configDir := filepath.Join(homeDir, ".config", appName)
		err := os.MkdirAll(configDir, 0700)
		if err != nil {
			logger.Debug().Err(err).Msg("error while creating directory in linux,darwin os")
			return fallback
		}
		historyPath = filepath.Join(configDir, "history")
	case "windows":
		appData := os.Getenv("AppData")
		configDir := filepath.Join(appData, appName)
		err := os.MkdirAll(configDir, 0700)
		if err != nil {
			logger.Debug().Err(err).Msg("error while creating directory in windows os")
			return fallback
		}
		historyPath = filepath.Join(configDir, "history")
	default:
		historyPath = filepath.Join(homeDir, "."+appName+"_history")
	}

	logger.Debug().Msgf("History Path is %s", historyPath)

	return historyPath
}
