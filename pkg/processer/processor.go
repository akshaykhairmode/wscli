package processer

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/akshaykhairmode/wscli/pkg/terminal"
	"github.com/akshaykhairmode/wscli/pkg/ws"

	"github.com/gorilla/websocket"
)

type Interactive struct {
	conn *websocket.Conn
	term *terminal.Term
}

func New(conn *websocket.Conn, term *terminal.Term) *Interactive {
	return &Interactive{
		conn: conn,
		term: term,
	}
}

func ProcessAsCmd(conn *websocket.Conn) {
	for _, cmd := range config.Flags.GetExecute() {
		ws.WriteToServer(conn, websocket.TextMessage, []byte(cmd))
	}

	defer func() {
		<-time.After(config.Flags.GetWait())
	}()

	if config.Flags.IsStdin() {
		go catchSignals(conn, nil)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			ws.WriteToServer(conn, websocket.TextMessage, scanner.Bytes())
		}
	}
}

func (i *Interactive) Process() {

	for _, cmd := range config.Flags.GetExecute() {
		ws.WriteToServer(i.conn, websocket.TextMessage, []byte(cmd))
	}

	i.term.AppendPrompt(fmt.Sprintf("(%s)»", truncateString(config.Flags.GetConnectURL(), 25)))

	i.term.OnMessage(func(line string) {
		switch {
		case shouldProcessCommand(line, "/flags"):
			log.Println(config.Flags.String())
		case shouldProcessCommand(line, "/ping"):
			getPingPongHandler(i.conn, line, websocket.PingMessage)()
		case shouldProcessCommand(line, "/pong"):
			getPingPongHandler(i.conn, line, websocket.PongMessage)()
		case shouldProcessCommand(line, "/close"):
			closeHandler(i.conn, line)
		case shouldProcessCommand(line, "/bfile"):
			sendBinaryFile(i.conn, line)
		default:
			ws.WriteToServer(i.conn, websocket.TextMessage, []byte(line))
		}
	})

}

func sendBinaryFile(conn *websocket.Conn, line string) {
	filePath := strings.TrimSpace(line[6:])
	if filePath == "" {
		log.Println("filepath is empty")
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("file open err : %s", err)
		return
	}

	fi, err := f.Stat()
	if err != nil {
		log.Printf("file stat err : %s", err)
		return
	}

	if fi.Size() > 50*1024*1024 {
		log.Println("file size is greater than 50mb")
		return
	}

	fileData, err := io.ReadAll(f)
	if err != nil {
		log.Printf("readall error : %s", err)
		return
	}

	ws.WriteToServer(conn, websocket.BinaryMessage, fileData)
	log.Println("file sent successfully")

}

func shouldProcessCommand(line, prefix string) bool {
	if config.Flags.IsSlash() && strings.HasPrefix(line, prefix) {
		return true
	}

	return false
}

func truncateString(s string, n int) string {
	r := []rune(s) // Convert to rune slice to handle Unicode correctly
	if len(r) > n {
		return string(r[:n]) + "..."
	}
	return s
}

func closeHandler(conn *websocket.Conn, line string) {
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
			logger.Err(err).Msg("write close error")
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

func catchSignals(conn *websocket.Conn, term *terminal.Term) {
	sigs := make(chan os.Signal, 2)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	logger.Debug().Msgf("received signal %s", <-sigs)

	if conn != nil {
		if err := conn.Close(); err != nil {
			logger.Debug().Err(err).Msg("error while closing connection")
		}
	}

	if term != nil {
		term.Close()
	}

	time.Sleep(300 * time.Millisecond)
	os.Exit(0)

}
