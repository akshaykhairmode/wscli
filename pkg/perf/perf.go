package perf

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/akshaykhairmode/wscli/pkg/ws"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

type Generator struct {
	config      config.Perf
	metric      *Metrics
	loadMessage messageGetter
	authMessage messageGetter
}

func New() (*Generator, error) {

	//If config file is passed, overwrite perf config.
	if config.Flags.Perf.ConfigPath != "" {
		cfgBytes, err := os.ReadFile(config.Flags.Perf.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("error while reading the config file : %w", err)
		}

		if err := yaml.Unmarshal(cfgBytes, &config.Flags.Perf); err != nil {
			return nil, fmt.Errorf("error while unmarshalling the config file : %w", err)
		}
	}

	if config.Flags.Perf.LogOutFile != "" {
		logger.Init(LogBuffer, fileFormatLevelFunc)
	} else {
		logger.Init(LogBuffer, nil)
	}

	if config.Flags.Perf.TotalConns <= 0 {
		return nil, fmt.Errorf("total number of connections are required")
	}

	lm, err := NewMessage(config.Flags.Perf.LoadMessage)
	if err != nil {
		return nil, fmt.Errorf("error while getting the load message : %w", err)
	}

	am, err := NewMessage(config.Flags.Perf.AuthMessage)
	if err != nil {
		return nil, fmt.Errorf("error while getting the auth message : %w", err)
	}

	logger.Info().Msgf("Config Loaded : %s", config.Flags.Perf)

	return &Generator{
		config:      config.Flags.Perf,
		metric:      NewMetrics(int64(config.Flags.Perf.TotalConns), config.Flags.Perf.LogOutFile),
		loadMessage: lm,
		authMessage: am,
	}, nil
}

func (g *Generator) Run(showTview bool) {

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-sigs
		g.metric.output.Stop()
		os.Exit(0)
	}()

	logger.Info().Msg("Started the load test")

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

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
	}()

	defer g.metric.printFinalMetrics()

	if showTview {
		g.metric.output.Start()
	} else {
		wg.Wait()

		select {}
	}

}

func (g *Generator) messgeReceiver(conn *websocket.Conn, wg *sync.WaitGroup, waitChan chan struct{}) {
	defer wg.Done()

	defer func() {
		waitChan <- struct{}{}
	}()

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
	defer g.metric.IncrDroppedConnections()

	//connect
	now := time.Now()
	conn, closef, _, err := ws.Connect()
	if err != nil {
		logger.Error().Err(err).Msg("error while connecting")
		return
	}
	defer g.metric.DecrActiveConnections()
	g.metric.IncrActiveConnections()
	g.metric.SetAvgConnectTime(time.Since(now))

	defer closef()

	waitChan := make(chan struct{}, 1)
	defer func() {
		<-waitChan
	}()

	//read messages
	wg.Add(1)
	go g.messgeReceiver(conn, wg, waitChan)

	//Wait for some time before sending auth.
	if g.config.WaitBeforeAuth > 0 {
		<-time.After(g.config.WaitBeforeAuth)
	}

	//send auth message
	if g.config.AuthMessage != "" {
		func() {
			msg, release := g.authMessage.Get(nil)
			defer release()

			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				logger.Error().Err(err).Msg("error while sending the auth message")
				return
			}
		}()

	}

	//wait for auth response.
	if g.config.WaitAfterAuth > 0 {
		<-time.After(g.config.WaitAfterAuth)
	}

	//if load message is empty then we dont send messages
	if g.config.LoadMessage == "" {
		return
	}

	seqCounter := uint64(0)

	lmFunc := func() {
		now := time.Now()

		msg, release := g.loadMessage.Get(Sequence{seqCounter})
		defer release()

		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			g.metric.IncrFailedMessages()
			logger.Err(err).Msg("error while sending the load message")
			return
		}

		g.metric.SetAvgMessageTime(time.Since(now))
		g.metric.IncrSentMessages()
		seqCounter++
	}

	//Send Load Mesasge only once when mi is 0.
	if g.config.MessageInterval <= 0 {
		lmFunc()
		return
	}

	//send load
	for range time.Tick(g.config.MessageInterval) {
		lmFunc()
	}
}

type Sequence struct {
	Seq uint64
}
