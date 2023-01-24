package orch

import (
	"context"
	"log"
	"time"
)

const DisconnectionPeriod = 5 * time.Second

type NodeConnection struct {
	Name            string
	Address         string
	LastInteraction time.Time
}

func DisconnectThread(orch *Orch, ctx context.Context, done func()) {
	ticker := time.NewTicker(DisconnectionPeriod)
	for {
		select {
		case <-ticker.C:
			// Kick any nodes that we haven't heard from in the last DisconnectionPeriod
			for name, n := range orch.Nodes {
				if n.LastInteraction.Before(time.Now().Add(-1 * DisconnectionPeriod)) {
					delete(orch.Nodes, name)
					log.Printf("Kicked out node: %s\n", name)
				}
			}
		case <-ctx.Done():
			return
		}
	}

}
