package orchestrator

import (
	"context"
	"sync"
)

type Orchestrator struct {
	Port  int
	Nodes map[string]*NodeConnection
}

func (orchestrator *Orchestrator) Run() error {
	if orchestrator.Nodes == nil {
		orchestrator.Nodes = make(map[string]*NodeConnection)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	done := func() {
		cancel()
		wg.Done()
	}

	wg.Add(1)
	go OServerThread(orchestrator, ctx, done)
	wg.Add(1)
	go DisconnectThread(orchestrator, ctx, done)

	wg.Wait()
	cancel()
	return nil
}
