// nolint
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var (
	upgrader = newUpgrader()
)

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	u.SetPingHandler(func(c *websocket.Conn, s string) {
		log.Println("Received Ping:", s)
		c.WriteMessage(websocket.PongMessage, []byte(s))
	})

	u.SetPongHandler(func(c *websocket.Conn, s string) {
		log.Println("Received Pong:", s)
		c.WriteMessage(websocket.PingMessage, []byte(s))
	})

	u.OnOpen(func(c *websocket.Conn) {
		// echo
		fmt.Println("OnOpen:", c.RemoteAddr().String())

		// go func() {
		// 	time.Sleep(5 * time.Second)
		// 	c.WriteClose(3008, "closing after 10s")
		// }()

	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {

		if string(data) == "zip" {
			zipData, err := zipStringToGzipBytes("hello, this is zipped data")
			if err != nil {
				log.Println(err)
				return
			}
			c.WriteMessage(websocket.BinaryMessage, zipData)
		}

		log.Println(messageType, string(data))
		// echo
		fmt.Println("OnMessage:", messageType, string(data))
		c.WriteMessage(messageType, data)
	})
	u.OnClose(func(c *websocket.Conn, err error) {
		fmt.Println("OnClose:", c.RemoteAddr().String(), err)
	})

	return u
}

func zipStringToGzipBytes(input string) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write([]byte(input))
	if err != nil {
		return nil, fmt.Errorf("failed to write gzip data: %w", err)
	}
	err = gz.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return b.Bytes(), nil
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Upgraded:", conn.RemoteAddr().String())
}

func main() {
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", onWebsocket)
	engine := nbhttp.NewEngine(nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{"localhost:8080"},
		MaxLoad:                 1000000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	})

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	engine.Shutdown(ctx)
}
