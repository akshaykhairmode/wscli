// nolint
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"flag"
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
		// log.Println("Received Ping:", s)
		c.WriteMessage(websocket.PongMessage, []byte(s))
	})

	u.SetPongHandler(func(c *websocket.Conn, s string) {
		// log.Println("Received Pong:", s)
		c.WriteMessage(websocket.PingMessage, []byte(s))
	})

	u.OnOpen(func(c *websocket.Conn) {
		fmt.Println("OnOpen:", c.RemoteAddr().String())
	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {

		if string(data) == "zip" {
			zipData, err := zipStringToGzipBytes("hello, this is zipped data")
			if err != nil {
				log.Println(err)
				return
			}
			// Use the standard BinaryMessage type to send the compressed data.
			c.WriteMessage(websocket.BinaryMessage, zipData)
		}

		log.Println(messageType, string(data))
		// echo
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
	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Log the error instead of panicking, which allows the server to keep running
		log.Printf("Websocket upgrade failed: %v", err)
		return
	}
}

func main() {
	// Define a flag for the Unix Domain Socket path
	udsPath := flag.String("uds", "", "Path to the Unix Domain Socket file (e.g., /tmp/ws.sock). If set, listens on UDS instead of TCP.")
	flag.Parse()

	var network string
	var addrs []string

	if *udsPath != "" {
		// --- Unix Domain Socket Setup ---
		network = "unix"
		addrs = []string{*udsPath}
		fmt.Printf("Starting WebSocket server on UDS: %s\n", *udsPath)

		// 1. Clean up stale socket file before starting (required if previous run crashed)
		if err := os.Remove(*udsPath); err != nil && !os.IsNotExist(err) {
			// Log as fatal since we cannot proceed if the file can't be removed
			log.Fatalf("Failed to clean up old UDS file %s: %v", *udsPath, err)
		}

		// 2. Clean up socket file on graceful exit
		defer func() {
			fmt.Printf("Cleaning up UDS file: %s\n", *udsPath)
			if err := os.Remove(*udsPath); err != nil && !os.IsNotExist(err) {
				log.Printf("Error cleaning up UDS file %s: %v", *udsPath, err)
			}
		}()

	} else {
		// --- TCP Default Setup ---
		network = "tcp"
		addrs = []string{":8080"}
		fmt.Printf("Starting WebSocket server on TCP: %s\n", addrs[0])
	}

	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", onWebsocket)

	engine := nbhttp.NewEngine(nbhttp.Config{
		Network:                 network, // Dynamically set network type
		Addrs:                   addrs,   // Dynamically set addresses
		MaxLoad:                 1000000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	})

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	// Graceful shutdown handling
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	engine.Shutdown(ctx)
}
