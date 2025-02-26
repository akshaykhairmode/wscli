package main

import (
	"fmt"
	"log"
	"os"

	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/global"
	"github.com/akshaykhairmode/wscli/pkg/logger"
	"github.com/akshaykhairmode/wscli/pkg/processer"
	"github.com/akshaykhairmode/wscli/pkg/terminal"
	"github.com/akshaykhairmode/wscli/pkg/ws"
)

var CLIVersion string

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
}

func main() {

	logger.Init()

	if config.Flags.IsShowVersion() {
		logger.Info().Msgf("CLI Version : %s", CLIVersion)
		return
	}

	conn, closeFunc, err := ws.Connect()
	if err != nil {
		logger.Fatal().Err(err).Msg("connect err")
	}

	defer closeFunc()

	if config.Flags.ShouldProcessAsCmd() {
		processer.ProcessAsCmd(conn)
		return
	}

	term, closef, wg := terminal.New()
	defer func() {
		if err := closef(); err != nil {
			logger.Debug().Err(err).Msg("error while closing readline")
		}
	}()

	go func() {
		global.WaitForStop()
		term.Close()
	}()

	processer.New(conn, term).Process()

	term.Reader(wg)

	fmt.Println()
}
