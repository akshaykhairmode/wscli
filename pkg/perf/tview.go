package perf

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	TotalConnections      = "Total"
	ActiveConnections     = "Active"
	DroppedConnections    = "Dropped"
	TotalSentMessages     = "M-Sent"
	TotalReceivedMessages = "M-Received"
	TotalFailedMessages   = "M-Failed"

	ConnectionMeanTime = "C-Mean"
	ConnectionP95Time  = "C-P95"
	ConnectionP99Time  = "C-P99"

	MessageMeanTime = "M-Mean"
	MessageP95Time  = "M-P95"
	MessageP99Time  = "M-P99"

	StartTime = "StartTime"
	Uptime    = "Uptime"
)

var headings = []string{
	TotalConnections,
	ActiveConnections,
	DroppedConnections,
	TotalSentMessages,
	TotalReceivedMessages,
	TotalFailedMessages,

	ConnectionMeanTime,
	ConnectionP95Time,
	ConnectionP99Time,

	MessageMeanTime,
	MessageP95Time,
	MessageP99Time,

	StartTime,
	Uptime,
}

// Latency coloring thresholds (milliseconds).
const (
	latencyGreenMs  = 10
	latencyOrangeMs = 100
)

const progressBarWidth = 20

// compactHeightThreshold is the minimum terminal height for the full three-pane
// layout. Below it we switch to a single scrollable metrics pane so the
// dashboard stays usable in short terminals (e.g. IDE terminal panels).
const compactHeightThreshold = 24

// Palette — a calm, cohesive "monitoring dashboard" theme. These are tview
// dynamic-color tags (W3C names or #RRGGBB hex).
const (
	colText   = "#f1f3f5" // primary text
	colMuted  = "#868e96" // labels / secondary
	colTrack  = "#495057" // progress bar track
	colGood   = "#40c057" // healthy / success
	colInfo   = "#4dabf7" // informational (uptime, rates)
	colWarn   = "#f59f00" // warning (ramping up, latency warn)
	colBad    = "#fa5252" // errors / failed
	colAccent = "#66d9e8" // app name + connections accent (cyan)
)

// Pane accent colors (tcell.Color for borders/titles).
var (
	accentHeader  = tcell.NewHexColor(0x74c0fc)
	accentConns   = tcell.NewHexColor(0x66d9e8)
	accentMsgs    = tcell.NewHexColor(0x69db7c)
	accentLatency = tcell.NewHexColor(0xffc078)
	accentLogs    = tcell.NewHexColor(0xff8787)
)

type Tview struct {
	app     *tview.Application
	header  *tview.TextView
	conns   *tview.TextView
	msgs    *tview.TextView
	latency *tview.TextView
	logs    *tview.TextView
	footer  *tview.TextView
	grid    *tview.Grid

	compactHeader  *tview.TextView
	compactFooter  *tview.TextView
	compactMetrics *tview.TextView
	compacted      bool

	autoScroll bool

	lastSent   int64
	lastRecv   int64
	lastFailed int64
	lastUpdate time.Time
}

func NewTview() *Tview {
	tviewApplication := tview.NewApplication()

	tviewHeader := tview.NewTextView()
	tviewHeader.SetDynamicColors(true)
	tviewHeader.SetTextAlign(tview.AlignCenter)
	tviewHeader.SetBorder(true)
	tviewHeader.SetTitle(" Status ")
	tviewHeader.SetTitleColor(accentHeader)
	tviewHeader.SetBorderColor(accentHeader)

	connsPane := newMetricPane(" Connections ", accentConns)
	msgsPane := newMetricPane(" Messages ", accentMsgs)
	latencyPane := newMetricPane(" Latency ", accentLatency)

	tviewLog := tview.NewTextView()
	tviewLog.SetDynamicColors(true)
	tviewLog.SetScrollable(true)
	tviewLog.SetWrap(true)
	tviewLog.SetBorder(true)
	tviewLog.SetTitleColor(accentLogs)
	tviewLog.SetBorderColor(accentLogs)

	tviewFooter := tview.NewTextView()
	tviewFooter.SetDynamicColors(true)
	tviewFooter.SetTextAlign(tview.AlignCenter)
	tviewFooter.SetBorder(true)
	tviewFooter.SetTitle(" Keys ")
	tviewFooter.SetTitleColor(accentHeader)
	tviewFooter.SetBorderColor(accentHeader)

	tv := &Tview{
		app:        tviewApplication,
		header:     tviewHeader,
		conns:      connsPane,
		msgs:       msgsPane,
		latency:    latencyPane,
		logs:       tviewLog,
		footer:     tviewFooter,
		autoScroll: true,
	}

	tviewLog.SetChangedFunc(func() {
		if tv.autoScroll {
			tv.logs.ScrollToEnd()
		}
	})

	tviewApplication.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyEsc || event.Rune() == 'q' || event.Rune() == 'Q':
			tviewApplication.Stop()
			return nil
		case event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown ||
			event.Key() == tcell.KeyPgUp || event.Key() == tcell.KeyPgDn:
			tv.setAutoScroll(false)
		case event.Key() == tcell.KeyEnd || event.Rune() == 'r' || event.Rune() == 'R':
			tv.setAutoScroll(true)
		}
		return event
	})

	metricsGrid := tview.NewGrid().
		SetRows(0, 0, 0).
		AddItem(connsPane, 0, 0, 1, 1, 0, 0, false).
		AddItem(msgsPane, 1, 0, 1, 1, 0, 0, false).
		AddItem(latencyPane, 2, 0, 1, 1, 0, 0, false)

	tv.grid = tview.NewGrid().
		SetRows(3, 0, 2).
		SetColumns(-2, -1).
		AddItem(tviewHeader, 0, 0, 1, 2, 0, 0, false).
		AddItem(metricsGrid, 1, 0, 1, 1, 0, 0, false).
		AddItem(tviewLog, 1, 1, 1, 1, 0, 30, true).
		AddItem(tviewFooter, 2, 0, 1, 2, 0, 0, false)

	compactHeader := tview.NewTextView()
	compactHeader.SetDynamicColors(true)
	compactHeader.SetTextAlign(tview.AlignCenter)

	compactFooter := tview.NewTextView()
	compactFooter.SetDynamicColors(true)
	compactFooter.SetTextAlign(tview.AlignCenter)

	compactMetrics := tview.NewTextView()
	compactMetrics.SetDynamicColors(true)
	compactMetrics.SetScrollable(true)
	compactMetrics.SetWrap(true)
	compactMetrics.SetBorder(true)
	compactMetrics.SetTitle(" Metrics ")
	compactMetrics.SetTitleColor(accentConns)
	compactMetrics.SetBorderColor(accentConns)

	tv.compactHeader = compactHeader
	tv.compactFooter = compactFooter
	tv.compactMetrics = compactMetrics

	tv.refreshBadges()

	return tv
}

func newMetricPane(title string, titleColor tcell.Color) *tview.TextView {
	pane := tview.NewTextView()
	pane.SetDynamicColors(true)
	pane.SetScrollable(false)
	pane.SetWrap(true)
	pane.SetBorder(true)
	pane.SetTitle(title)
	pane.SetTitleColor(titleColor)
	pane.SetBorderColor(titleColor)
	return pane
}

func (tv *Tview) Start() {
	tv.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		if !tv.compacted {
			_, height := screen.Size()
			if height > 0 && height < compactHeightThreshold {
				tv.compacted = true
				tv.buildCompactGrid()
			}
		}
		return false
	})

	if err := tv.app.SetRoot(tv.grid, true).Run(); err != nil {
		logger.Err(err).Send()
	}
}

// buildCompactGrid replaces the contents of the root grid with the compact
// single-scrollable-metrics layout. It mutates the grid in place (no SetRoot)
// so it can be called from the before-draw handler without deadlocking.
func (tv *Tview) buildCompactGrid() {
	tv.grid.
		Clear().
		SetRows(1, 0, 1).
		SetColumns(-2, -1).
		AddItem(tv.compactHeader, 0, 0, 1, 2, 0, 0, false).
		AddItem(tv.compactMetrics, 1, 0, 1, 1, 0, 0, false).
		AddItem(tv.logs, 1, 1, 1, 1, 0, 30, true).
		AddItem(tv.compactFooter, 2, 0, 1, 2, 0, 0, false)
}

func (tv *Tview) Stop() {
	tv.app.Stop()
}

func (tv *Tview) setAutoScroll(autoScroll bool) {
	tv.autoScroll = autoScroll
	if autoScroll {
		tv.logs.ScrollToEnd()
	}
	tv.refreshBadges()
}

func (tv *Tview) refreshBadges() {
	if tv.autoScroll {
		tv.logs.SetTitle(" Errors / Logs [AUTO] ")
	} else {
		tv.logs.SetTitle(" Errors / Logs [PAUSED] ")
	}
}

func (tv *Tview) UpdateTableAndLogs(data map[string]string, errors *errMsg) {
	tv.app.QueueUpdateDraw(func() {
		tv.header.SetText(renderHeader(data))
		tv.conns.SetText(renderConnections(data))
		tv.msgs.SetText(tv.renderMessages(data))
		tv.latency.SetText(renderLatency(data))
		tv.footer.SetText(renderFooter(tv.autoScroll))

		tv.compactHeader.SetText(renderHeader(data))
		tv.compactFooter.SetText(renderFooter(tv.autoScroll))
		tv.compactMetrics.SetText(strings.Join([]string{
			renderConnections(data),
			tv.renderMessages(data),
			renderLatency(data),
		}, "\n\n"))

		builder := strings.Builder{}
		errors.ForEach(func(data map[string]int, order []string) {
			for _, v := range order {
				if data[v] > 1 {
					builder.WriteString(fmt.Sprintf("[%s]%s[white] [%s](%d)[white]\n", colBad, v, colMuted, data[v]))
				} else {
					builder.WriteString(fmt.Sprintf("[%s]%s[white]\n", colWarn, v))
				}
			}
		})

		if builder.Len() == 0 {
			builder.WriteString(fmt.Sprintf("[%s]No errors recorded yet.[white]\n", colGood))
		}

		tv.logs.SetText(builder.String())
		tv.refreshBadges()
	})
}

func renderHeader(data map[string]string) string {
	active := parseLeadingInt(data[ActiveConnections])
	total := parseLeadingInt(data[TotalConnections])

	status := "Running"
	statusColor := colGood
	if active < total {
		status = "Ramping up"
		statusColor = colWarn
	}

	return fmt.Sprintf(
		"[::b][%s]wscli[white] - Load Testing    [%s]Uptime: %s[white]    [%s]Active: %d/%d[white]    [%s]%s[white]",
		colAccent, colInfo, data[Uptime], colGood, active, total, statusColor, status,
	)
}

func tag(s string) string {
	return fmt.Sprintf("[::b][%s]%s[white][::B]", colMuted, s)
}

func renderConnections(data map[string]string) string {
	active := parseLeadingInt(data[ActiveConnections])
	total := parseLeadingInt(data[TotalConnections])

	droppedColor := colMuted
	if parseLeadingInt(data[DroppedConnections]) > 0 {
		droppedColor = colBad
	}

	return fmt.Sprintf(
		"%s   %s\n%s  %s\n%s %s\n%s",
		tag("Total:"), data[TotalConnections],
		tag("Active:"), fmt.Sprintf("[%s]%s[white]", colGood, data[ActiveConnections]),
		tag("Dropped:"), fmt.Sprintf("[%s]%s[white]", droppedColor, data[DroppedConnections]),
		progressBar(active, total, progressBarWidth),
	)
}

func (tv *Tview) renderMessages(data map[string]string) string {
	sent := parseLeadingInt(data[TotalSentMessages])
	recv := parseLeadingInt(data[TotalReceivedMessages])
	failed := parseLeadingInt(data[TotalFailedMessages])

	now := time.Now()
	elapsed := now.Sub(tv.lastUpdate)
	rateSent := formatRate(sent-tv.lastSent, elapsed)
	rateRecv := formatRate(recv-tv.lastRecv, elapsed)
	rateFailed := formatRate(failed-tv.lastFailed, elapsed)

	tv.lastSent = sent
	tv.lastRecv = recv
	tv.lastFailed = failed
	tv.lastUpdate = now

	failedColor := colGood
	if failed > 0 {
		failedColor = colBad
	}

	return fmt.Sprintf(
		"%s     %s  [%s]%s[white]\n%s %s  [%s]%s[white]\n%s   [%s]%s[white]  [%s]%s[white]",
		tag("Sent:"), formatInt(sent), colInfo, rateSent,
		tag("Received:"), formatInt(recv), colInfo, rateRecv,
		tag("Failed:"), failedColor, formatInt(failed), failedColor, rateFailed,
	)
}

func renderLatency(data map[string]string) string {
	connMean := colorizeDuration(data[ConnectionMeanTime])
	connP95 := colorizeDuration(data[ConnectionP95Time])
	connP99 := colorizeDuration(data[ConnectionP99Time])
	msgMean := colorizeDuration(data[MessageMeanTime])
	msgP95 := colorizeDuration(data[MessageP95Time])
	msgP99 := colorizeDuration(data[MessageP99Time])

	return fmt.Sprintf(
		"%s mean %s  p95 %s  p99 %s\n%s mean %s  p95 %s  p99 %s",
		tag("Connect:"), connMean, connP95, connP99,
		tag("Message:"), msgMean, msgP95, msgP99,
	)
}

func colorizeDuration(s string) string {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Sprintf("[%s]n/a[white]", colMuted)
	}
	return fmt.Sprintf("[%s]%s[white]", latencyColor(d), s)
}

func latencyColor(d time.Duration) string {
	ms := d.Milliseconds()
	switch {
	case ms < latencyGreenMs:
		return colGood
	case ms < latencyOrangeMs:
		return colWarn
	default:
		return colBad
	}
}

func renderFooter(autoScroll bool) string {
	mode := fmt.Sprintf("[%s]AUTO[white]", colGood)
	if !autoScroll {
		mode = fmt.Sprintf("[%s]PAUSED[white]", colBad)
	}
	return fmt.Sprintf(
		"[%s][q][white] quit    [%s][↑/↓/PgUp/PgDn][white] scroll logs    [%s][r/End][white] follow    scroll: %s",
		colInfo, colInfo, colInfo, mode,
	)
}

func progressBar(active, total int64, width int) string {
	if width <= 0 {
		width = progressBarWidth
	}

	pct := float64(0)
	if total > 0 {
		pct = float64(active) / float64(total)
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	filled := int(math.Round(float64(width) * pct))
	if filled > width {
		filled = width
	}

	fillColor := colAccent
	switch {
	case pct >= 0.9:
		fillColor = colGood
	case pct >= 0.5:
		fillColor = colInfo
	}

	return fmt.Sprintf(
		"[%s]%s[%s]%s[white] [%s]%.1f%%[white]",
		fillColor,
		strings.Repeat("█", filled),
		colTrack,
		strings.Repeat("░", width-filled),
		colMuted,
		pct*100,
	)
}

func formatRate(delta int64, elapsed time.Duration) string {
	if elapsed <= 0 {
		return "+0/s"
	}
	return fmt.Sprintf("+%s/s", formatInt(int64(math.Round(float64(delta)/elapsed.Seconds()))))
}

func formatInt(n int64) string {
	in := strconv.FormatInt(n, 10)
	if len(in) <= 3 {
		return in
	}

	var b strings.Builder
	start := len(in) % 3
	if start > 0 {
		b.WriteString(in[:start])
		b.WriteByte(',')
	}
	for i := start; i < len(in); i += 3 {
		b.WriteString(in[i : i+3])
		if i+3 < len(in) {
			b.WriteByte(',')
		}
	}
	return b.String()
}

func parseLeadingInt(s string) int64 {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0
	}
	n, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0
	}
	return n
}
