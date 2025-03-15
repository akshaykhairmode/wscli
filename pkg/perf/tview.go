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
	app   *tview.Application
	table *tview.Table
	logs  *tview.TextView
	grid  *tview.Grid
}

func NewTview() *Tview {

	tviewApplication := tview.NewApplication()

	tviewTable := tview.NewTable().SetBorders(true)

	tviewLog := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	tviewLog.SetBorder(true)
	tviewLog.SetTitle(" wscli - Load Testing ")
	tviewLog.SetTitleColor(tcell.ColorBlue)

	tviewGrid := tview.NewGrid().
		SetRows(0, 0, 0).
		AddItem(tviewTable, 0, 0, 1, 1, 0, 0, false).
		AddItem(tviewLog, 1, 0, 2, 1, 0, 0, true).
		SetBorders(false)

	tviewLog.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown || event.Key() == tcell.KeyPgUp || event.Key() == tcell.KeyPgDn {
			logAutoScroll = false
		}
		return event
	})

	for col, h := range headings {
		tviewTable.SetCell(0, col, tview.NewTableCell(h).SetTextColor(tcell.ColorBlue).SetAlign(tview.AlignCenter))
	}

	return &Tview{
		app:   tviewApplication,
		table: tviewTable,
		logs:  tviewLog,
		grid:  tviewGrid,
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

func (tv *Tview) UpdateTableAndLogs(data []string, errors errMsg) {

	tv.app.QueueUpdateDraw(func() {
		updateTable(tv.table, data)

		errors, order := errors.Get()
		builder := strings.Builder{}
		for _, v := range order {
			if errors[v] > 1 {
				builder.WriteString(fmt.Sprintf("%s [blue](%d)[white]\n", v, errors[v]))
			} else {
				builder.WriteString(fmt.Sprintf("%s\n", v))
			}
		}

		tv.logs.SetText(builder.String())

		if logAutoScroll {
			tv.logs.ScrollToEnd().ScrollToHighlight()
		}
	})

}

func updateTable(table *tview.Table, values []string) {
	for col, val := range values {
		cell := tview.NewTableCell(val).SetAlign(tview.AlignCenter)
		if col == 2 {
			cell.SetTextColor(tcell.ColorRed)
		} else {
			cell.SetTextColor(tcell.ColorGreen)
		}
		table.SetCell(1, col, cell)
	}
}
