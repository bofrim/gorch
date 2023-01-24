package orch

import (
	"context"
	"sync"
)

type Orch struct {
	Port  int
	Nodes map[string]NodeConnection
}

func (orch *Orch) Run() error {
	if orch.Nodes == nil {
		orch.Nodes = make(map[string]NodeConnection)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	done := func() {
		cancel()
		wg.Done()
	}

	wg.Add(1)
	go ServerThread(orch, ctx, done)
	wg.Add(1)
	go DisconnectThread(orch, ctx, done)

	// Finish
	wg.Wait()
	cancel()
	return nil
}
