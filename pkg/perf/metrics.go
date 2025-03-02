package perf

import (
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"text/tabwriter"
	"time"
)

type Metrics struct {
	activeConnections     *atomic.Int64
	totalSentMessages     *atomic.Int64
	totalReceivedMessages *atomic.Int64
	failedMessages        *atomic.Int64

	avgConnectTimeMS *atomic.Int64
	avgMessageTimeMS *atomic.Int64

	totalConns int64

	tw *tabwriter.Writer
}

func NewMetrics(totalConns int64) *Metrics {
	m := &Metrics{
		activeConnections:     &atomic.Int64{},
		totalSentMessages:     &atomic.Int64{},
		totalReceivedMessages: &atomic.Int64{},
		failedMessages:        &atomic.Int64{},
		avgConnectTimeMS:      &atomic.Int64{},
		avgMessageTimeMS:      &atomic.Int64{},
		totalConns:            totalConns,
		tw:                    tabwriter.NewWriter(os.Stdout, 0, 6, 1, ' ', tabwriter.TabIndent),
	}

	go m.printMetrics()

	return m
}

func MSToNS(milliseconds int64) time.Duration {
	return time.Duration(milliseconds * 1000000)
}

func (m *Metrics) printMetrics() {

	heading := []any{
		"TotConns",
		"ActConns",
		"TotSent",
		"TotRecvd",
		"TotFailed",
		"AvgConnTime",
		"AvgMsgTime",
	}

	fmt.Println()
	for range time.Tick(time.Second) {
		total := m.totalConns
		active := m.activeConnections.Load()
		sent := m.totalSentMessages.Load()
		received := m.totalReceivedMessages.Load()
		failed := m.failedMessages.Load()

		avgConnectTime := time.Duration(0)
		if active > 0 {
			avgConnectTime = MSToNS(m.avgConnectTimeMS.Load()) / time.Duration(active)
		}

		avgMessageTime := time.Duration(0)
		if sent > 0 {
			avgMessageTime = MSToNS(m.avgMessageTimeMS.Load()) / time.Duration(sent)
		}

		m.print([][]any{
			heading,
			{total, active, sent, received, failed, avgConnectTime.Round(time.Millisecond), avgMessageTime.Round(time.Millisecond)},
		})
	}
}

func moveCursorToStart() {
	if runtime.GOOS == "windows" {
		fmt.Print("\r")
	} else {
		fmt.Print("\033[2F")
	}
}

func (m *Metrics) print(data [][]any) {
	moveCursorToStart()
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
}

func (m *Metrics) IncrActiveConnections() {
	m.activeConnections.Add(1)
}

func (m *Metrics) DecrActiveConnections() {
	m.activeConnections.Add(-1)
}

func (m *Metrics) IncrSentMessages() {
	m.totalSentMessages.Add(1)
}

func (m *Metrics) IncrFailedMessages() {
	m.failedMessages.Add(1)
}

func (m *Metrics) IncrReceivedMessages() {
	m.totalReceivedMessages.Add(1)
}

func (m *Metrics) SetAvgConnectTime(dur time.Duration) {
	m.avgConnectTimeMS.Add(dur.Milliseconds())
}

func (m *Metrics) SetAvgMessageTime(dur time.Duration) {
	m.avgMessageTimeMS.Add(dur.Milliseconds())
}
