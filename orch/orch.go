package orch

import (
	"context"
	"sync"
)

type Orch struct {
	Port  int
	nodes []string
}

func (orch *Orch) Run() error {
	var wg sync.WaitGroup
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	done := func() {
		cancel()
		wg.Done()
	}

	go ServerThread(orch, ctx, done)

	// Finish
	wg.Wait()
	cancel()
	return nil
}
