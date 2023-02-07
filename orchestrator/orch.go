package orchestrator

import (
	"context"
	"sync"

	"github.com/bofrim/gorch/utils"
	"golang.org/x/exp/slog"
)

type Orchestrator struct {
	Port     int
	Nodes    map[string]*NodeConnection
	LogFile  string
	CertPath string
}

func (orchestrator *Orchestrator) Run() (err error) {
	if orchestrator.Nodes == nil {
		orchestrator.Nodes = make(map[string]*NodeConnection)
	}
	var logger *slog.Logger
	var closeFn func()
	if logger, closeFn, err = utils.SetupLogging(orchestrator.LogFile); err != nil {
		return err
	}
	defer closeFn()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	done := func() {
		cancel()
		wg.Done()
	}

	wg.Add(1)
	go OServerThread(orchestrator, ctx, logger, done)
	wg.Add(1)
	go DisconnectThread(orchestrator, ctx, logger, done)

	logger.Info("Orchestrator is up and running!")
	wg.Wait()
	cancel()
	return nil
}
