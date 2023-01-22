package node

import (
	"context"
	"time"
)

type ClientState int

const ClientPollPeriod = 1 * time.Second

const (
	Polling ClientState = iota
	Registered
)

func ClientThread(n Node, ctx context.Context, done func()) {
	defer done()

	ticker := time.NewTicker(ClientPollPeriod)
	for {
		select {
		case <-ticker.C:
			if n.OrchConnState == Polling {
				// Attempt to Register

			}
		case <-ctx.Done():
			return
		}
	}
}
