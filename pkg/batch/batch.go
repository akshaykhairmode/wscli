package batch

import (
	"bufio"
	"fmt"
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

func Process(conn *websocket.Conn, cfg config.Config, rl *readline.Instance) {

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go ws.ReadMessages(cfg, conn, wg, rl)

	for _, cmd := range cfg.Execute {
		ws.WriteToServer(conn, cmd)
	}

	<-time.After(cfg.Wait)

	if len(cfg.Execute) > 0 && cfg.Wait > 0 {
		return
	}

	if cfg.Stdin {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			ws.WriteToServer(conn, scanner.Text())
		}
		return
	}

	log.SetOutput(rl.Stderr())
	log.SetFlags(0)

	log.Println(ws.GreenColor("Connected"))

	rl.SetPrompt(GetPrompt(cfg, fmt.Sprintf("(%s)»", truncateString(cfg.ConnectURL, 25))))
	rl.Refresh()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				conn.Close()
				break
			} else {
				continue
			}
		}
		if err != nil {
			logger.GlobalLogger.Fatal().Err(err).Msg("error while doing readline")
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

}

func GetPrompt(cfg config.Config, str string) string {
	if cfg.NoColor {
		return str
	}

	return fmt.Sprintf("\033[31m%s\033[0m ", str)
}

func truncateString(s string, n int) string {
	r := []rune(s) // Convert to rune slice to handle Unicode correctly
	if len(r) > n {
		return string(r[:n]) + "..."
	}
	return s
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
