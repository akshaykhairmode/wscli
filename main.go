package main

import (
	"log"
	"os"
	"sync"

	"github.com/akshaykhairmode/wscli/pkg/batch"
	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/interactive"
	"github.com/akshaykhairmode/wscli/pkg/logger"
)

func main() {

	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	cfg := config.Get()

	logger.Init(cfg)

	wg := &sync.WaitGroup{}

	if cfg.IsInteractive {
		interactive.Process(cfg, wg)
		return
	}

	batch.Process(cfg)
}
