package interactive

import (
	"io"
	"log"
	"strings"
	"sync"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/ws"
	"github.com/chzyer/readline"
)

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("/connect"),
	readline.PcItem("/exit"),
	readline.PcItem("/ping"),
	readline.PcItem("/pong"),
	readline.PcItem("/wait"),
	readline.PcItem("/help"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func Process(cfg config.Config, wg *sync.WaitGroup) {

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
		Listener:            nil,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch {
		case line == "/ping":

		case line == "/help":
			usage(l.Stderr())
			continue
		case strings.HasPrefix(line, "/exit"):
			return
		case strings.HasPrefix(line, "/connect"):
			line = strings.TrimSpace(line[8:])
			conn, _, err := ws.Connect(line, false, cfg.Headers, cfg.Auth)
			if err != nil {
				log.Println("connect err,", err)
			} else {
				go ws.ReadMessages(cfg, conn, wg)
				log.Println("Connected to ", line)
			}
		default:
			// ws.WriteToServer(WSConn, line)
		}
	}
}
