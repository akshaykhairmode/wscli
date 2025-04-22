package perf

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/logger"
)

type FileOutput struct {
	f *os.File
	w *tabwriter.Writer
}

func NewFileOutput(path string) *FileOutput {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0644)
	if err != nil {
		logger.Fatal().Err(err).Msg("error while opening the output file")
	}

	w := tabwriter.NewWriter(f, 0, 0, 2, ' ', tabwriter.AlignRight|tabwriter.Debug)

	out := &FileOutput{
		f: f,
		w: w,
	}

	return out
}

func (fo *FileOutput) tWrite(data string) {
	_, err := fo.w.Write(fmt.Appendf(nil, "%s\t", data))
	if err != nil {
		logger.Err(err).Msg("error while writing to tabwriter")
	}
}

func (fo *FileOutput) UpdateTableAndLogs(data []string, errors *errMsg) {

	//Stats
	for _, heading := range headings {
		fo.tWrite(heading)
	}

	fo.tWrite("\n")

	for _, value := range data {
		fo.tWrite(value)
	}

	fo.tWrite("\n")

	defer fo.w.Flush()

	//print errors
	if len(data) <= 0 {
		return
	}

	now := time.Now().Format(timeFormat)

	errors.ForEach(func(data map[string]int, order []string) {
		for _, v := range order {
			if data[v] > 1 {
				fmt.Fprintf(fo.w, "%s %s (%d)\n", now, v, data[v])
			} else {
				fmt.Fprintf(fo.w, "%s %s\n", now, v)
			}
		}
	})

	fmt.Fprintln(fo.w, "----------------------------------------------------------------------------------")
}

func (fo *FileOutput) Start() {}

func (fo *FileOutput) Stop() {}

var fileFormatLevelFunc = func(i any) string {
	level := strings.ToUpper(i.(string))
	return level
}
