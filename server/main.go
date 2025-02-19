// nolint
package main

import (
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
	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {

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
