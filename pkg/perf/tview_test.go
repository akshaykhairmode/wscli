package perf

import (
	"strings"
	"testing"
	"time"

	"github.com/rcrowley/go-metrics"
)

func TestProgressBar(t *testing.T) {
	bar := progressBar(50, 100, 10)
	if got := strings.Count(bar, "█"); got != 5 {
		t.Errorf("expected 5 filled blocks, got %d (%q)", got, bar)
	}
	if got := strings.Count(bar, "░"); got != 5 {
		t.Errorf("expected 5 empty blocks, got %d (%q)", got, bar)
	}
	if !strings.Contains(bar, "50.0%") {
		t.Errorf("expected 50.0%% in bar, got %q", bar)
	}

	bar = progressBar(0, 0, 10)
	if got := strings.Count(bar, "█"); got != 0 {
		t.Errorf("expected no filled blocks for zero total, got %d (%q)", got, bar)
	}

	bar = progressBar(200, 100, 10)
	if got := strings.Count(bar, "█"); got != 10 {
		t.Errorf("expected 10 filled blocks when over 100%%, got %d (%q)", got, bar)
	}

	bar = progressBar(50, 100, 0)
	if !strings.Contains(bar, "%") {
		t.Errorf("expected progress bar with default width, got %q", bar)
	}
}

func TestFormatRate(t *testing.T) {
	if got := formatRate(100, time.Second); got != "+100/s" {
		t.Errorf("formatRate(100, 1s) = %q, want +100/s", got)
	}
	if got := formatRate(1234, time.Second); got != "+1,234/s" {
		t.Errorf("formatRate(1234, 1s) = %q, want +1,234/s", got)
	}
	if got := formatRate(50, 2*time.Second); got != "+25/s" {
		t.Errorf("formatRate(50, 2s) = %q, want +25/s", got)
	}
	if got := formatRate(0, 0); got != "+0/s" {
		t.Errorf("formatRate(0, 0) = %q, want +0/s", got)
	}
}

func TestFormatInt(t *testing.T) {
	cases := map[int64]string{
		0:       "0",
		5:       "5",
		999:     "999",
		1000:    "1,000",
		12345:   "12,345",
		1000000: "1,000,000",
		-12345:  "-12,345",
	}
	for in, want := range cases {
		if got := formatInt(in); got != want {
			t.Errorf("formatInt(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestParseLeadingInt(t *testing.T) {
	cases := map[string]int64{
		"675 (67.50%)": 675,
		"1000":         1000,
		"":             0,
		"abc":          0,
	}
	for in, want := range cases {
		if got := parseLeadingInt(in); got != want {
			t.Errorf("parseLeadingInt(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestGetTable(t *testing.T) {
	m := &Metrics{
		activeConnections:     metrics.NewCounter(),
		droppedConnections:    metrics.NewCounter(),
		totalSentMessages:     metrics.NewCounter(),
		totalReceivedMessages: metrics.NewCounter(),
		failedMessages:        metrics.NewCounter(),
		connectTime:           metrics.NewTimer(),
		messageTime:           metrics.NewTimer(),
		totalConns:            100,
		startTime:             time.Now(),
		startTimeStr:          "1:00:00 PM",
	}
	m.activeConnections.Inc(25)
	m.totalSentMessages.Inc(50)

	data := m.getTable()
	for _, heading := range headings {
		if _, ok := data[heading]; !ok {
			t.Errorf("getTable output is missing heading %q", heading)
		}
	}

	if got := data[TotalConnections]; got != "100" {
		t.Errorf("TotalConnections = %q, want 100", got)
	}
	if got := data[ActiveConnections]; !strings.Contains(got, "25") {
		t.Errorf("ActiveConnections = %q, want it to contain 25", got)
	}
	if got := data[TotalSentMessages]; got != "50" {
		t.Errorf("TotalSentMessages = %q, want 50", got)
	}
}

func TestRenderHeader(t *testing.T) {
	data := map[string]string{
		TotalConnections:  "100",
		ActiveConnections: "50 (50.00%)",
		Uptime:            "5s",
	}

	header := renderHeader(data)
	if !strings.Contains(header, "Active: 50/100") {
		t.Errorf("expected Active: 50/100 in header, got %q", header)
	}
	if !strings.Contains(header, "Ramping up") {
		t.Errorf("expected 'Ramping up' status while active < total, got %q", header)
	}

	data[ActiveConnections] = "100 (100.00%)"
	header = renderHeader(data)
	if !strings.Contains(header, "Running") {
		t.Errorf("expected 'Running' status when active == total, got %q", header)
	}
}
