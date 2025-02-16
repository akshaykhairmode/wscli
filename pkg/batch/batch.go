package batch

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/akshaykhairmode/wscli/pkg/ws"
	"github.com/chzyer/readline"
	"github.com/gorilla/websocket"
)

func Process(cfg config.Config) {

	conn, closeFunc, err := ws.Connect(cfg)
	if err != nil {
		logger.GlobalLogger.Fatal().Err(err).Msg("error while connecting")
	}

	defer closeFunc()

	prompt := "\033[31m»\033[0m "
	if cfg.NoColor {
		prompt = "» "
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     "/tmp/readline_n.tmp",
		AutoComplete:    normalCompleter,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
		// FuncFilterInputRune: filterInput,
	})
	if err != nil {
		logger.GlobalLogger.Fatal().Err(err).Msg("error while creating readline object")
	}
	defer l.Close()
	l.CaptureExitSignal()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go ws.ReadMessages(cfg, conn, wg, l)

	if cfg.Execute != "" {
		ws.WriteToServer(conn, cfg.Execute)
	}

	<-time.After(cfg.Wait)

	if cfg.Execute != "" && cfg.Wait > 0 {
		return
	}

	if cfg.Stdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			ws.WriteToServer(conn, scanner.Text())
		}
		return
	}

	log.SetOutput(l.Stderr())
	log.SetFlags(0)

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		}
		if err != nil {
			log.Fatal(err)
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		switch {
		case line == "/exit" || line == "exit":
			return
		case cfg.IsSlash && strings.HasPrefix(line, "/ping"):
			getPingPongHandler(conn, line, websocket.PingMessage)()
		case cfg.IsSlash && strings.HasPrefix(line, "/pong"):
			getPingPongHandler(conn, line, websocket.PongMessage)()
		case cfg.IsSlash && strings.HasPrefix(line, "/close"):
			closeHandler(line, conn)
		default:
			ws.WriteToServer(conn, line)
		}

	}

	conn.Close()

	wg.Wait()

}

func closeHandler(line string, conn *websocket.Conn) {
	str := strings.TrimSpace(line[6:])
	if len(str) > 0 {
		spl := strings.Split(str, " ")
		if len(spl) < 2 {
			log.Println("invalid close message, close message must have code and reason")
			return
		}

		closeCode, err := strconv.Atoi(spl[0])
		if err != nil {
			log.Println("invalid close code, must be a number")
			return
		}

		reason := strings.TrimSpace(strings.Join(spl[1:], " "))

		if err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, reason), time.Now().Add(3*time.Second)); err != nil {
			logger.GlobalLogger.Err(err).Msg("write close error")
		}
	}
}

func getPingPongHandler(conn *websocket.Conn, line string, mt int) func() {
	return func() {
		str := strings.TrimSpace(line[5:])
		if err := conn.WriteControl(mt, []byte(str), time.Now().Add(3*time.Second)); err != nil {
			log.Println(err)
		}
	}
}

var normalCompleter = readline.NewPrefixCompleter(
	readline.PcItem("/exit"),
	readline.PcItem("/ping"),
	readline.PcItem("/pong"),
	readline.PcItem("/close"),
)
