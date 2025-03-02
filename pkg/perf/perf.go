package perf

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/akshaykhairmode/wscli/pkg/ws"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Generator struct {
	config config.Perf
	templ  *template.Template
	metric *Metrics
}

func New(config config.Perf) (*Generator, error) {

	if config.TotalConns <= 0 {
		return nil, fmt.Errorf("total number of connections are required")
	}

	tmpl := template.New("parse").Funcs(funcMap)
	if err := parseTemplate(tmpl, config.LoadMessage); err != nil {
		return nil, fmt.Errorf("error while parsing the load message : %w", err)
	}

	if err := parseTemplate(tmpl, config.AuthMessage); err != nil {
		return nil, fmt.Errorf("error while parsing the auth message : %w", err)
	}

	return &Generator{
		config: config,
		templ:  tmpl,
		metric: NewMetrics(int64(config.TotalConns)),
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
		if err := conn.WriteMessage(websocket.TextMessage, g.executeTemplate(g.config.AuthMessage)); err != nil {
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
		if err := conn.WriteMessage(websocket.TextMessage, g.executeTemplate(g.config.LoadMessage)); err != nil {
			g.metric.IncrFailedMessages()
			logger.Debug().Err(err).Msg("error while sending the load message")
			return
		}
		g.metric.SetAvgMessageTime(time.Since(now))
		g.metric.IncrSentMessages()
	}
}

func (g *Generator) executeTemplate(data string) []byte {
	buf := bytes.NewBuffer(nil)
	err := g.templ.Execute(buf, data)
	if err != nil {
		logger.Error().Err(err).Msgf("error while executing the template : %s", data)
		return nil
	}
	return buf.Bytes()
}

var funcMap = template.FuncMap{
	"RandomNumber":       randomInt,
	"RandomUUID":         randomUUID,
	"RandomAlphaNumeric": randomAlphaNumeric,
}

const alphaNumericChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomAlphaNumeric(length ...int) string {
	l := 10
	if len(length) > 0 {
		l = length[0]
	}

	b := make([]byte, l)
	for i := range b {
		b[i] = alphaNumericChars[rand.Intn(len(alphaNumericChars))]
	}

	return string(b)
}

func randomInt(max ...int) int {
	if len(max) <= 0 {
		return rand.Intn(10000)
	}

	return rand.Intn(max[0])
}

func randomUUID() string {
	guid := uuid.New()
	return guid.String()
}

func parseTemplate(tmpl *template.Template, str string) error {

	if str == "" {
		return nil
	}

	if _, err := tmpl.Parse(str); err != nil {
		return fmt.Errorf("error while parsing the template : %w", err)
	}

	return nil

}
