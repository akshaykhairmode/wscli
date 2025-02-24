package terminal

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/akshaykhairmode/wscli/pkg/ws"
	"github.com/chzyer/readline"
)

type Term struct {
	rl        *readline.Instance
	onMessage func(line string)
}

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
	if config.Flags.IsNoColor() {
		return str
	}

	return fmt.Sprintf("\033[31m%s\033[0m ", str)
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

type CloseFunc func() error

func New() (*Term, CloseFunc, *sync.WaitGroup) {

	if config.Flags.IsStdin() {
		return &Term{}, func() error { return nil }, &sync.WaitGroup{}
	}

	rl, err := readline.NewEx(getDefaultConfig())
	if err != nil {
		logger.GlobalLogger.Fatal().Err(err).Msg("error while creating readline object")
	}

	if !config.Flags.IsStdin() {
		rl.CaptureExitSignal()
	}

	wg := &sync.WaitGroup{}

	term := &Term{rl: rl}

	wg.Add(1)
	go term.reader(wg)

	log.SetOutput(term.GetOutLoc())
	log.SetFlags(0)

	log.Println(ws.GreenColor("Connected"))

	return term, rl.Close, wg
}

func (t *Term) Close() error {
	return t.rl.Close()
}

func (t *Term) AppendPrompt(prompt string) {
	t.rl.SetPrompt(getPrompt(prompt))
	t.rl.Refresh()
}

func (t *Term) GetOutLoc() io.Writer {
	return t.rl.Stderr()
}

func (t *Term) reader(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		line, err := t.rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				return
			} else {
				continue
			}
		}
		if err != nil {
			logger.GlobalLogger.Debug().Err(err).Msg("error while doing next line in terminal")
			return
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if t.onMessage != nil {
			t.onMessage(line)
		}

	}
}

func (t *Term) OnMessage(f func(line string)) {
	t.onMessage = f
}
