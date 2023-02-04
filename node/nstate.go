package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bofrim/gorch/orchestrator"
	"golang.org/x/exp/slog"
)

const NodePollPeriod time.Duration = orchestrator.DisconnectStaleNodePeriod / 2
const QuickPollThreshold int = 25
const NodeQuickPollPeriod time.Duration = NodePollPeriod / time.Duration(QuickPollThreshold)

type NodeCommState string

const (
	Polling       NodeCommState = "polling"
	QuickPolling  NodeCommState = "quick-polling"
	Registered    NodeCommState = "registered"
	Disconnecting NodeCommState = "disconnecting"
	Disconnected  NodeCommState = "disconnected"
)

type NodeState struct {
	commState NodeCommState
	pollCount int
}

func (ns *NodeState) ChangeState(state NodeCommState) {
	slog.Default().Info(
		"Change node comm state",
		slog.String("old", string(ns.commState)),
		slog.String("new", string(state)),
	)
	ns.commState = state
}

func NodeStateThread(n *Node, ctx context.Context, logger *slog.Logger, done func()) {
	defer done()
	n.nodeState.commState = Polling
	ticker := time.NewTicker(NodePollPeriod)

	// Attempt to register at startup to avoid waiting for the first period to elapse
	if err := register(n.OrchAddr, n.Name, n.ServerAddr, n.ServerPort); err == nil {
		logger.Debug("Start-up registration.", slog.String("node", n.Name))
		n.nodeState.commState = Registered
	}

	for {
		select {
		case <-ticker.C:
			switch n.nodeState.commState {
			case QuickPolling:
				n.nodeState.pollCount++
				if n.nodeState.pollCount > QuickPollThreshold {
					n.nodeState.pollCount = 0
					n.nodeState.ChangeState(Polling)
					ticker.Reset(NodePollPeriod)
				}
				fallthrough
			case Polling:
				if err := register(n.OrchAddr, n.Name, n.ServerAddr, n.ServerPort); err == nil {
					n.nodeState.ChangeState(Registered)
				}
			case Registered:
				if err := ping(n.OrchAddr, n.Name); err != nil {
					n.nodeState.ChangeState(QuickPolling)
					ticker.Reset(NodeQuickPollPeriod)
				}
			case Disconnecting:
				_ = disconnect(n.OrchAddr, n.Name)
				n.nodeState.ChangeState(Disconnected)
				return
			case Disconnected:
				return
			default:
				n.nodeState.ChangeState(Disconnecting)
			}

		case <-ctx.Done():
			return
		}
	}
}

func register(orchAddr, nodeName string, nodeAddr string, nodePort int) error {
	// Register with the orchestrator
	url := fmt.Sprintf("http://%s/register/", orchAddr)
	data := orchestrator.NodeRegistration{
		NodeName: nodeName,
		NodeAddr: nodeAddr,
		NodePort: nodePort,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("request not OK: %d", resp.StatusCode)
		log.Println(err.Error())
		return err
	}
	log.Printf("Registered as [%s (%s:%d)] to [%s]", nodeName, nodeAddr, nodePort, orchAddr)
	return nil
}

func ping(addr, name string) error {
	url := fmt.Sprintf("http://%s/ping/%s", addr, name)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad request: %d", resp.StatusCode)
	}
	return nil
}

func disconnect(addr, name string) error {

	return nil
}
