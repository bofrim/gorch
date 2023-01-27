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
)

type NodeState struct {
	commState NodeCommState
	pollCount int
}
type NodeCommState int

const NodePollPeriod time.Duration = orchestrator.DisconnectStaleNodePeriod / 2
const QuickPollThreshold int = 25
const NodeQuickPollPeriod time.Duration = time.Duration(int(NodePollPeriod) / QuickPollThreshold)

const (
	Polling NodeCommState = iota
	QuickPolling
	Registered
	Disconnecting
	Disconnected
)

func NodeStateThread(n *Node, ctx context.Context, done func()) {
	defer done()
	n.nodeState.commState = Polling
	ticker := time.NewTicker(NodePollPeriod)

	// Attempt to register at startup to avoid waiting for the first period to elapse
	if err := register(n.OrchAddr, n.Name, n.ServerAddr, n.ServerPort); err == nil {
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
					n.nodeState.commState = Polling
					ticker.Reset(NodePollPeriod)
				}
				fallthrough
			case Polling:
				if err := register(n.OrchAddr, n.Name, n.ServerAddr, n.ServerPort); err == nil {
					n.nodeState.commState = Registered
				}
			case Registered:
				if err := ping(n.OrchAddr, n.Name); err != nil {
					log.Println("Bad ping. Going back to polling")
					n.nodeState.commState = QuickPolling
					ticker.Reset(NodeQuickPollPeriod)
				}
			case Disconnecting:
				log.Println("Disconnecting.")
				_ = disconnect(n.OrchAddr, n.Name)
				n.nodeState.commState = Disconnected
				return
			case Disconnected:
				log.Println("Disconnected.")
				return
			default:
				n.nodeState.commState = Disconnecting
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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Bad request")
		return fmt.Errorf("request not OK: %d", resp.StatusCode)
	}
	log.Println("Registered")
	return nil
}

func ping(addr, name string) error {
	url := fmt.Sprintf("http://%s/ping/%s", addr, name)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
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
