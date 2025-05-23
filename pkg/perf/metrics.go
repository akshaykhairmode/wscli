package perf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/rcrowley/go-metrics"
)

type Metrics struct {
	activeConnections     metrics.Counter
	droppedConnections    metrics.Counter
	totalSentMessages     metrics.Counter
	totalReceivedMessages metrics.Counter
	failedMessages        metrics.Counter

	connectTime metrics.Timer
	messageTime metrics.Timer

	totalConns int64

	errors *errMsg
	output Printer

	startTime    time.Time
	startTimeStr string
}

type Printer interface {
	UpdateTableAndLogs(data []string, errors *errMsg)
	Start()
	Stop()
}

type errMsg struct {
	data  map[string]int
	order []string
	mux   *sync.RWMutex
}

func (em *errMsg) Add(msg string) {
	em.mux.Lock()
	defer em.mux.Unlock()

	if _, ok := em.data[msg]; !ok {
		em.order = append(em.order, msg)
	}

	em.data[msg]++
}

func (em *errMsg) ForEach(f func(data map[string]int, order []string)) {
	em.mux.RLock()
	defer em.mux.RUnlock()
	f(em.data, em.order)
}

func NewMetrics(totalConns int64, out string) *Metrics {

	var output Printer

	if out == "" {
		output = NewTview()
	} else {
		output = NewFileOutput(out)
	}

	now := time.Now()

	m := &Metrics{
		activeConnections:     metrics.NewCounter(),
		droppedConnections:    metrics.NewCounter(),
		totalSentMessages:     metrics.NewCounter(),
		totalReceivedMessages: metrics.NewCounter(),
		failedMessages:        metrics.NewCounter(),
		connectTime:           metrics.NewTimer(),
		messageTime:           metrics.NewTimer(),
		totalConns:            totalConns,
		errors: &errMsg{
			data: make(map[string]int),
			mux:  &sync.RWMutex{},
		},
		output:       output,
		startTime:    now,
		startTimeStr: now.Format(timeFormat),
	}

	metrics.MustRegister("active_connections", m.activeConnections)
	metrics.MustRegister("dropped_connections", m.droppedConnections)
	metrics.MustRegister("total_sent", m.totalSentMessages)
	metrics.MustRegister("total_received", m.totalReceivedMessages)
	metrics.MustRegister("total_failed", m.failedMessages)
	metrics.MustRegister("connection_time", m.connectTime)
	metrics.MustRegister("message_time", m.messageTime)

	go m.printMetrics()

	return m
}

type customBufferedLogger struct {
	buf *bytes.Buffer
	mux *sync.RWMutex
}

func (l *customBufferedLogger) Write(p []byte) (n int, err error) {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.buf.Write(p)
}

func (l *customBufferedLogger) Read(buf []byte) (int, error) {
	l.mux.RLock()
	defer l.mux.RUnlock()
	return l.buf.Read(buf)
}

var LogBuffer = &customBufferedLogger{
	buf: &bytes.Buffer{},
	mux: &sync.RWMutex{},
}

var re = regexp.MustCompile(`^\d{2}:\d{2}:\d{2}.\d{3} `)

func stripTimeFromLog(log string) string {
	return re.ReplaceAllString(log, "")
}

func (m *Metrics) printMetrics() {

	go func() {
		r := bufio.NewReader(LogBuffer)

		for {
			data, _, err := r.ReadLine()
			if err == io.EOF {
				time.Sleep(200 * time.Millisecond)
				continue
			}

			if len(data) <= 0 {
				continue
			}

			str := stripTimeFromLog(string(data))

			m.errors.Add(str)
		}

	}()

	go func() {
		m.output.UpdateTableAndLogs(m.getTable(headings), m.errors)

		for range time.Tick(config.Flags.PrintOutputInterval) {
			m.output.UpdateTableAndLogs(m.getTable(headings), m.errors)
		}
	}()

}

func (m *Metrics) printFinalMetrics() {

	values := m.getTable(headings)
	for index, heading := range headings {
		fmt.Printf("%s,%s\n", heading, values[index])
	}

}

const (
	timeFormat = "3:04:05 PM"
	p95        = 0.95
	p99        = 0.99
)

func (m *Metrics) getTable(heading []string) []string {

	final := []string{}

	connectTime := m.connectTime.Snapshot()
	messageTime := m.messageTime.Snapshot()

	for _, val := range heading {

		switch val {
		case TotalConnections:
			final = append(final, intToString(m.totalConns))
		case ActiveConnections:
			final = append(final, calculatePercentage(m.activeConnections.Count(), m.totalConns))
		case DroppedConnections:
			final = append(final, calculatePercentage(m.droppedConnections.Count(), m.totalConns))
		case TotalSentMessages:
			final = append(final, intToString(m.totalSentMessages.Count()))
		case TotalReceivedMessages:
			final = append(final, intToString(m.totalReceivedMessages.Count()))
		case TotalFailedMessages:
			final = append(final, intToString(m.failedMessages.Count()))
		case ConnectionMeanTime:
			final = append(final, durToString(connectTime.Mean()))
		case ConnectionP95Time:
			final = append(final, durToString(connectTime.Percentile(p95)))
		case ConnectionP99Time:
			final = append(final, durToString(connectTime.Percentile(p99)))
		case MessageMeanTime:
			final = append(final, durToString(messageTime.Mean()))
		case MessageP95Time:
			final = append(final, durToString(messageTime.Percentile(p95)))
		case MessageP99Time:
			final = append(final, durToString(messageTime.Percentile(p99)))
		case StartTime:
			final = append(final, m.startTimeStr)
		case Uptime:
			final = append(final, time.Since(m.startTime).Round(time.Second).String())
		}
	}

	return final

}

func intToString(i int64) string {
	return strconv.Itoa(int(i))
}

func durToString(f float64) string {
	return time.Duration(f).Round(time.Millisecond).String()
}

func calculatePercentage(value, total int64) string {
	if total == 0 {
		return "0.00%"
	}

	percentage := (float64(value) / float64(total)) * 100

	return fmt.Sprintf("%d (%.2f%%)", value, percentage)

}

func (m *Metrics) IncrDroppedConnections() {
	m.droppedConnections.Inc(1)
}

func (m *Metrics) IncrActiveConnections() {
	m.activeConnections.Inc(1)
}

func (m *Metrics) DecrActiveConnections() {
	m.activeConnections.Dec(1)
}

func (m *Metrics) IncrSentMessages() {
	m.totalSentMessages.Inc(1)
}

func (m *Metrics) IncrFailedMessages() {
	m.failedMessages.Inc(1)
}

func (m *Metrics) IncrReceivedMessages() {
	m.totalReceivedMessages.Inc(1)
}

func (m *Metrics) SetAvgConnectTime(dur time.Duration) {
	m.connectTime.Update(dur)
}

func (m *Metrics) SetAvgMessageTime(dur time.Duration) {
	m.messageTime.Update(dur)
}
