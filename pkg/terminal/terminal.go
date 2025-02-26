package terminal

import (
	"io"
	"log"
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

type CloseFunc func() error

func New() (*Term, CloseFunc, *sync.WaitGroup) {

	if config.Flags.IsStdin() {
		return &Term{}, func() error { return nil }, &sync.WaitGroup{}
	}

	rl, err := readline.NewEx(getDefaultConfig())
	if err != nil {
		logger.Fatal().Err(err).Msg("error while creating readline object")
	}

	if !config.Flags.IsStdin() {
		rl.CaptureExitSignal()
	}

	wg := &sync.WaitGroup{}

	term := &Term{rl: rl}

	log.SetOutput(term.GetOutLoc())
	log.SetFlags(0)

	log.Println(ws.GreenColor("Connected"))

	return term, rl.Close, wg
}

func (t *Term) Close() {
	if err := t.rl.Close(); err != nil {
		logger.Debug().Err(err).Msg("error while closing terminal")
	}
}

func (t *Term) AppendPrompt(prompt string) {
	t.rl.SetPrompt(getPrompt(prompt))
	t.rl.Refresh()
}

func (t *Term) GetOutLoc() io.Writer {
	return t.rl.Stderr()
}

func (t *Term) Reader(wg *sync.WaitGroup) {

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
			logger.Debug().Err(err).Msg("error while doing next line in terminal")
			return
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if line == "/exit" || line == "exit" {
			return
		}

		if t.onMessage != nil {
			t.onMessage(line)
		}

	}
}

func (t *Term) OnMessage(f func(line string)) {
	t.onMessage = f
}
