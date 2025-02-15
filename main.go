package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/ws"
	"github.com/chzyer/readline"
	"github.com/gorilla/websocket"
)

func main() {

	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	cfg := config.Get()

	log.Printf("%+v", cfg)

	if cfg.IsInteractive {
		processInteractive(cfg)
		return
	}

	conn, closeFunc, err := ws.Connect(cfg.ConnectURL, cfg.Response, cfg.Headers, cfg.Auth)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFunc()

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	wg.Add(1)
	go ws.ReadMessages(cfg, conn, wg)

	if cfg.Execute != "" {
		err = conn.WriteMessage(websocket.TextMessage, []byte(cfg.Execute))
		if err != nil {
			log.Println(err)
		}
	}

	<-time.After(cfg.Wait)

	if cfg.Stdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if scanner.Err() != nil {
				log.Println("scan err", scanner.Err())
				return
			}
			err := conn.WriteMessage(websocket.TextMessage, scanner.Bytes())
			if err != nil {
				log.Println("write err", err)
			}
		}
		return
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline_n.tmp",
		AutoComplete:    normalCompleter,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()

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
		case cfg.IsSlash && line == "/close":
			closeHandler(line, conn)
		default:
			err = conn.WriteMessage(websocket.TextMessage, []byte(line))
			if err != nil {
				log.Println(err)
			}
		}

	}

}

func closeHandler(line string, conn *websocket.Conn) {
	str := strings.TrimSpace(line[6:])
	if len(str) > 0 {
		spl := strings.Split(str, " ")
		if len(spl) != 2 {
			log.Println("invalid close message, close message must have code and reason")
			return
		}

		reason := strings.TrimSpace(spl[1])

		closeCode, err := strconv.Atoi(spl[0])
		if err != nil {
			log.Println("invalid close code, must be a number")
			return
		}

		if err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, reason), time.Now().Add(3*time.Second)); err != nil {
			log.Println(err)
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
