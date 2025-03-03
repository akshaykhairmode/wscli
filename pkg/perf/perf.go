package perf

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/akshaykhairmode/wscli/pkg/ws"
	"github.com/gorilla/websocket"
)

type Generator struct {
	config      config.Perf
	metric      *Metrics
	loadMessage MessageGetter
	authMessage MessageGetter
}

func New(config config.Perf) (*Generator, error) {

	if config.TotalConns <= 0 {
		return nil, fmt.Errorf("total number of connections are required")
	}

	lm, err := NewMessage(config.LoadMessage)
	if err != nil {
		return nil, fmt.Errorf("error while getting the load message : %w", err)
	}

	am, err := NewMessage(config.AuthMessage)
	if err != nil {
		return nil, fmt.Errorf("error while getting the auth message : %w", err)
	}

	return &Generator{
		config:      config,
		metric:      NewMetrics(int64(config.TotalConns)),
		loadMessage: lm,
		authMessage: am,
	}, nil
}

func (g *Generator) Run() {

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-sigs
		log.Printf("\n\n")
		os.Exit(0)
	}()

	wg := &sync.WaitGroup{}

	total := g.config.TotalConns

loop:
	for range time.Tick(time.Second) {

		for range g.config.RampUpConnsPerSecond {

			if total <= 0 {
				break loop
			}

			wg.Add(1)
			go g.processConnection(wg)
			total--
		}

	}

	wg.Wait()
}

func (g *Generator) messgeReceiver(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			logger.Err(err).Msg("error while reading the message")
			return
		}

		if len(data) <= 0 {
			continue
		}

		g.metric.IncrReceivedMessages()
	}
}

func (g *Generator) processConnection(wg *sync.WaitGroup) {
	defer wg.Done()
	defer g.metric.DecrActiveConnections()

	//connect
	now := time.Now()
	conn, closef, _, err := ws.Connect()
	if err != nil {
		logger.Error().Err(err).Msg("error while connecting")
		return
	}
	g.metric.IncrActiveConnections()
	g.metric.SetAvgConnectTime(time.Since(now))

	defer closef()

	//read messages
	wg.Add(1)
	go g.messgeReceiver(conn, wg)

	//send auth message
	if g.config.AuthMessage != "" {
		if err := conn.WriteMessage(websocket.TextMessage, g.authMessage.Get()); err != nil {
			logger.Error().Err(err).Msg("error while sending the auth message")
			return
		}
	}

	//wait for auth
	if g.config.WaitAfterAuth > 0 {
		<-time.After(g.config.WaitAfterAuth)
	}

	//if load message is empty then we dont send messages
	if g.config.LoadMessage == "" {
		return
	}

	//send load
	for range time.Tick(time.Second / time.Duration(g.config.MessagePerSecond)) {
		now := time.Now()
		if err := conn.WriteMessage(websocket.TextMessage, g.loadMessage.Get()); err != nil {
			g.metric.IncrFailedMessages()
			logger.Debug().Err(err).Msg("error while sending the load message")
			return
		}
		g.metric.SetAvgMessageTime(time.Since(now))
		g.metric.IncrSentMessages()
	}
}
