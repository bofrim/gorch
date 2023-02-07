package orchestrator

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/exp/slog"
)

const DisconnectStaleNodePeriod = 10 * time.Second

type NodeConnection struct {
	Name            string
	Address         string
	Port            int
	LastInteraction time.Time
}

func (nc *NodeConnection) RequestAction(actionName string, reqBody []byte) ([]byte, error) {
	url := fmt.Sprintf("https://%s:%d/action/%s", nc.Address, nc.Port, actionName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad request: %s\n", resp.Status)
		return nil, fmt.Errorf("request not OK: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}

func (nc *NodeConnection) GetRequest(reqBody []byte, path string) ([]byte, error) {
	url := fmt.Sprintf("https://%s:%d/%s", nc.Address, nc.Port, path)
	log.Printf("Get request for: %s", url)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad request: %s\n", resp.Status)
		return nil, fmt.Errorf("request not OK: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
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
