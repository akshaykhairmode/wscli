package perf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"text/template"
	"time"

	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/google/uuid"
)

type MessageGetter interface {
	Get() []byte
}

type File struct {
	reader   *bufio.Reader
	f        *os.File
	dataChan chan []byte
}

const dataChanSize = 1000

func NewMessage(fpath string) (MessageGetter, error) {

	isFile, size := isFilePath(fpath)
	if !isFile {
		return NewDefaultMessageGetter(fpath)
	}

	f, err := os.Open(fpath)
	if err != nil {
		return nil, fmt.Errorf("error while opening file : %w", err)
	}

	mg := &File{
		f:        f,
		dataChan: make(chan []byte, dataChanSize),
	}

	//If file size is less than equals to 10mb we will store in memory.
	if size <= 1024*1024*10 {
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("error while reading file : %w", err)
		}
		mg.reader = bufio.NewReader(bytes.NewReader(data))
		logger.Debug().Msgf("Loading File in memory")
	} else {
		mg.reader = bufio.NewReader(f)
	}

	go mg.readWorker()
	go mg.logWorker()

	return mg, nil
}

func (m *File) Get() []byte {
	return <-m.dataChan
}

func (m *File) logWorker() {
	for range time.Tick(time.Second) {
		if len(m.dataChan) < dataChanSize {
			logger.Debug().Msgf("Buffer Length is : %d", len(m.dataChan))
		}
	}
}

func (m *File) readWorker() {

	for {
		data, err := m.reader.ReadBytes('\n')
		if err == nil {
			m.dataChan <- bytes.TrimSuffix(data, []byte("\n"))
			continue
		}

		if err == io.EOF {
			m.f.Seek(0, 0)
			m.reader.Reset(m.f)
			continue
		}

		logger.Debug().Err(err).Msg("error while reading the file")
	}

}

func isFilePath(path string) (bool, int64) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, 0
	}

	if fileInfo.IsDir() {
		return false, 0
	}

	return true, fileInfo.Size()
}

type DefaultMessageGetter struct {
	msg  []byte
	tmpl *template.Template
}

func NewDefaultMessageGetter(msg string) (MessageGetter, error) {

	tmpl := template.New("parse").Funcs(funcMap)
	if err := parseTemplate(tmpl, msg); err != nil {
		return nil, fmt.Errorf("error while parsing the template : %s : %w", msg, err)
	}

	return &DefaultMessageGetter{
		msg:  []byte(msg),
		tmpl: tmpl,
	}, nil
}

func (m *DefaultMessageGetter) Get() []byte {
	buf := bytes.NewBuffer(nil)
	err := m.tmpl.Execute(buf, m.msg)
	if err != nil {
		logger.Error().Err(err).Msgf("error while executing the template : %s", m.msg)
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
