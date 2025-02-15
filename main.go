package main

import (
	"sync"

	"github.com/akshaykhairmode/wscli/pkg/batch"
	"github.com/akshaykhairmode/wscli/pkg/config"
	"github.com/akshaykhairmode/wscli/pkg/interactive"
	"github.com/akshaykhairmode/wscli/pkg/logger"
)

func main() {

	cfg := config.Get()

	logger.Init(cfg)

	wg := &sync.WaitGroup{}

	if cfg.IsInteractive {
		interactive.Process(cfg, wg)
		return
	}

	batch.Process(cfg, wg)

	wg.Wait()
}
