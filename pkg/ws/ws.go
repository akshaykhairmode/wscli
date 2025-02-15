package ws

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

func Connect(uri string, respHeader bool, cfgHeaders []string, auth string) (*websocket.Conn, func(), error) {

	closeFunc := func() {}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, closeFunc, fmt.Errorf("error while passing the url : %w", err)
	}

	headers := http.Header{}
	for _, h := range cfgHeaders {
		headSpl := strings.Split(h, ":")
		if len(headSpl) != 2 {
			return nil, closeFunc, fmt.Errorf("invalid header : %s", h)
		}
		headers.Set(headSpl[0], headSpl[1])
	}

	if auth != "" {
		headers.Set("Authorization", basicAuth(auth))
	}

	c, resp, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return nil, closeFunc, fmt.Errorf("dial err : %w", err)
	}

	if respHeader {
		for k, v := range resp.Header {
			log.Println(k, v)
		}

	}

	closeFunc = func() {
		c.Close()
	}

	return c, closeFunc, nil
}

func basicAuth(auth string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

var blueColor = color.New(color.FgBlue).SprintfFunc()
var greenColor = color.New(color.FgGreen).SprintfFunc()

func ReadMessages(cfg config.Config, conn *websocket.Conn, wg *sync.WaitGroup) {

	defer wg.Done()

	fn := func(what string) func(appData string) error {
		return func(appData string) error {
			if cfg.ShowPingPong {
				log.Println(blueColor("received %s (data: %s)", what, appData))
				// blueColor.Fprintf(log.Writer(), "received %s (data: %s)\n", what, appData)
			}
			return nil
		}
	}

	conn.SetPingHandler(fn("ping"))
	conn.SetPongHandler(fn("pong"))

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		switch mt {
		case websocket.TextMessage:
			log.Println(greenColor("< %s", string(message)))
		case websocket.BinaryMessage:
			log.Println(hex.EncodeToString(message))
		case websocket.CloseMessage:
			log.Println("received close message", message)
			return
		}

	}

}

func WriteToServer(conn *websocket.Conn, message string) {

	if conn == nil {
		log.Println("Connection is nil")
	} else {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Println("write:", err)
		}
	}

}
