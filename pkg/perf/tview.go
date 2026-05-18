package perf

import (
	"fmt"
	"strings"

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

var logAutoScroll = true

func init() {

}

type Tview struct {
	app    *tview.Application
	header *tview.TextView
	table  *tview.Table
	stats  *tview.TextView
	logs   *tview.TextView
	grid   *tview.Grid
}

func NewTview() *Tview {
	tviewApplication := tview.NewApplication()

	tviewHeader := tview.NewTextView()
	tviewHeader.SetDynamicColors(true)
	tviewHeader.SetTextAlign(tview.AlignCenter)
	tviewHeader.SetText("[::b][skyblue]wscli[white] - Load Testing    [darkcyan]Press ↑/↓ to scroll logs, Esc to quit[white]")
	tviewHeader.SetBorder(true)
	tviewHeader.SetTitle(" Status ")
	tviewHeader.SetTitleColor(tcell.ColorBlue)

	tviewTable := tview.NewTable()
	tviewTable.SetBorders(true)
	tviewTable.SetFixed(1, 0)
	tviewTable.SetSelectable(false, false)
	tviewTable.SetBorder(true)
	tviewTable.SetTitle(" Metrics ")
	tviewTable.SetTitleColor(tcell.ColorGreen)

	for col, h := range headings {
		tviewTable.SetCell(0, col, tview.NewTableCell(h).
			SetTextColor(tcell.ColorDarkKhaki).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold))
	}

	tviewStats := tview.NewTextView()
	tviewStats.SetDynamicColors(true)
	tviewStats.SetScrollable(false)
	tviewStats.SetWrap(true)
	tviewStats.SetBorder(true)
	tviewStats.SetTitle(" Summary ")
	tviewStats.SetTitleColor(tcell.ColorYellow)

	tviewLog := tview.NewTextView()
	tviewLog.SetDynamicColors(true)
	tviewLog.SetScrollable(true)
	tviewLog.SetWrap(true)
	tviewLog.SetBorder(true)
	tviewLog.SetTitle(" Errors / Logs ")
	tviewLog.SetTitleColor(tcell.ColorRed)

	tviewLog.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown || event.Key() == tcell.KeyPgUp || event.Key() == tcell.KeyPgDn {
			logAutoScroll = false
		}
		return event
	})

	tviewApplication.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' || event.Rune() == 'Q' {
			tviewApplication.Stop()
			return nil
		}
		return event
	})

	tviewGrid := tview.NewGrid().
		SetRows(3, 0, 0).
		SetColumns(0, 30).
		AddItem(tviewHeader, 0, 0, 1, 2, 0, 0, false).
		AddItem(tviewTable, 1, 0, 1, 1, 0, 0, false).
		AddItem(tviewStats, 1, 1, 1, 1, 0, 0, false).
		AddItem(tviewLog, 2, 0, 1, 2, 0, 0, true).
		SetBorders(false)

	return &Tview{
		app:    tviewApplication,
		header: tviewHeader,
		table:  tviewTable,
		stats:  tviewStats,
		logs:   tviewLog,
		grid:   tviewGrid,
	}
}

func (tv *Tview) Start() {
	if err := tv.app.SetRoot(tv.grid, true).Run(); err != nil {
		logger.Err(err).Send()
	}
}

func (tv *Tview) Stop() {
	tv.app.Stop()
}

func (tv *Tview) UpdateTableAndLogs(data []string, errors *errMsg) {
	tv.app.QueueUpdateDraw(func() {
		updateTable(tv.table, data)
		tv.stats.SetText(buildSummary(data))

		builder := strings.Builder{}
		errors.ForEach(func(data map[string]int, order []string) {
			for _, v := range order {
				if data[v] > 1 {
					builder.WriteString(fmt.Sprintf("[red]%s[white] [grey](%d)[white]\n", v, data[v]))
				} else {
					builder.WriteString(fmt.Sprintf("[yellow]%s[white]\n", v))
				}
			}
		})

		if builder.Len() == 0 {
			builder.WriteString("[green]No errors recorded yet.[white]\n")
		}

		tv.logs.SetText(builder.String())

		if logAutoScroll {
			tv.logs.ScrollToEnd().ScrollToHighlight()
		}
	})
}

func buildSummary(data []string) string {
	if len(data) < 14 {
		return ""
	}

	return fmt.Sprintf(
		"[::b]Connections:[white] Total %s  [green]Active %s[white]  [red]Dropped %s[white]\n[::b]Messages:[white] Sent %s  Recv %s  Fail %s\n[::b]Latency:[white] mean %s  p95 %s  p99 %s\n[::b]Timing:[white] start %s  uptime %s",
		data[0],
		data[1],
		data[2],
		data[3],
		data[4],
		data[5],
		data[6],
		data[7],
		data[8],
		data[12],
		data[13],
	)
}

func updateTable(table *tview.Table, values []string) {
	for col, val := range values {
		cell := tview.NewTableCell(val).SetAlign(tview.AlignCenter)
		switch col {
		case 2, 5:
			cell.SetTextColor(tcell.ColorRed)
		case 6, 7, 8, 9, 10, 11:
			cell.SetTextColor(tcell.ColorOrange)
		default:
			cell.SetTextColor(tcell.ColorGreen)
		}
		table.SetCell(1, col, cell)
	}
}
