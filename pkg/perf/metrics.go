package perf

import (
	"fmt"
	"os"
	"runtime"
	"text/tabwriter"
	"time"

	"github.com/rcrowley/go-metrics"
)

type Metrics struct {
	activeConnections     metrics.Counter
	totalSentMessages     metrics.Counter
	totalReceivedMessages metrics.Counter
	failedMessages        metrics.Counter

	connectTime metrics.Timer
	messageTime metrics.Timer

	totalConns int64

	tw *tabwriter.Writer
}

func NewMetrics(totalConns int64) *Metrics {
	m := &Metrics{
		activeConnections:     metrics.NewCounter(),
		totalSentMessages:     metrics.NewCounter(),
		totalReceivedMessages: metrics.NewCounter(),
		failedMessages:        metrics.NewCounter(),
		connectTime:           metrics.NewTimer(),
		messageTime:           metrics.NewTimer(),
		totalConns:            totalConns,
		tw:                    tabwriter.NewWriter(os.Stdout, 0, 6, 1, ' ', tabwriter.Debug|tabwriter.AlignRight),
	}

	metrics.MustRegister("active_connections", m.activeConnections)
	metrics.MustRegister("total_sent", m.totalSentMessages)
	metrics.MustRegister("total_received", m.totalReceivedMessages)
	metrics.MustRegister("total_failed", m.failedMessages)
	metrics.MustRegister("connection_time", m.connectTime)
	metrics.MustRegister("message_time", m.messageTime)

	go m.printMetrics()

	return m
}

const (
	TotalConnections      = "Total"
	ActiveConnections     = "Active"
	TotalSentMessages     = "Sent"
	TotalReceivedMessages = "Received"
	TotalFailedMessages   = "Failed"

	ConnectionMeanTime = "C_Mean"
	ConnectionP95Time  = "C_P95"
	ConnectionP99Time  = "C_P99"

	MessageMeanTime = "M_Mean"
	MessageP95Time  = "M_P95"
	MessageP99Time  = "M_P99"
)

func (m *Metrics) printMetrics() {

	heading := []any{
		TotalConnections,
		ActiveConnections,
		TotalSentMessages,
		TotalReceivedMessages,
		TotalFailedMessages,

		ConnectionMeanTime,
		ConnectionP95Time,
		ConnectionP99Time,

		MessageMeanTime,
		MessageP95Time,
		MessageP99Time,
	}

	for range time.Tick(time.Second) {

		final := []any{}

		for _, head := range heading {
			val := head.(string)

			var out any

			switch val {
			case TotalConnections:
				out = m.totalConns
			case ActiveConnections:
				out = m.activeConnections.Count()
			case TotalSentMessages:
				out = m.totalSentMessages.Count()
			case TotalReceivedMessages:
				out = m.totalReceivedMessages.Count()
			case TotalFailedMessages:
				out = m.failedMessages.Count()
			case ConnectionMeanTime:
				out = time.Duration(m.connectTime.Snapshot().Mean()).Round(time.Millisecond)
			case ConnectionP95Time:
				out = time.Duration(m.connectTime.Snapshot().Percentile(0.95)).Round(time.Millisecond)
			case ConnectionP99Time:
				out = time.Duration(m.connectTime.Snapshot().Percentile(0.99)).Round(time.Millisecond)
			case MessageMeanTime:
				out = time.Duration(m.messageTime.Snapshot().Mean()).Round(time.Millisecond)
			case MessageP95Time:
				out = time.Duration(m.messageTime.Snapshot().Percentile(0.95)).Round(time.Millisecond)
			case MessageP99Time:
				out = time.Duration(m.messageTime.Snapshot().Percentile(0.99)).Round(time.Millisecond)
			}

			final = append(final, out)

		}

		m.print([][]any{heading, final})
	}
}

func moveCursorToStart(length int) {
	for i := 0; i < length; i++ {
		if runtime.GOOS == "windows" {
			fmt.Print("\r")
		} else {
			fmt.Print("\033[F")
		}
	}
}

func (m *Metrics) print(data [][]any) {
	for _, row := range data {
		for i, cell := range row {
			fmt.Fprintf(m.tw, "%v", cell)
			if i < len(row)-1 {
				fmt.Fprintf(m.tw, "\t")
			}
		}
		fmt.Fprintf(m.tw, "\n")
	}
	m.tw.Flush()
	moveCursorToStart(len(data))
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
