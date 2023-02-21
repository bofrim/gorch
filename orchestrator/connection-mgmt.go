package orchestrator

import (
	"context"
	"time"

	"golang.org/x/exp/slog"
)

const DisconnectStaleNodePeriod = 10 * time.Second

type NodeConnection struct {
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	Port            int       `json:"port"`
	LastInteraction time.Time `json:"last_interaction"`
}

func DisconnectThread(orchestrator *Orchestrator, ctx context.Context, logger *slog.Logger, done func()) {
	ticker := time.NewTicker(DisconnectStaleNodePeriod)
	for {
		select {
		case <-ticker.C:
			// Kick any nodes that we haven't heard from in the last DisconnectStaleNodePeriod
			for name, n := range orchestrator.Nodes {
				if n.LastInteraction.Before(time.Now().Add(-1 * DisconnectStaleNodePeriod)) {
					delete(orchestrator.Nodes, name)
					logger.Info("Stale node.",
						slog.String("node", name),
						slog.Int("num_nodes", len(orchestrator.Nodes)),
					)
				}
			}
		case <-ctx.Done():
			return
		}
	}

}
