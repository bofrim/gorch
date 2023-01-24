package node

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ClientState int

const ClientPollPeriod = 1 * time.Second

const (
	Polling ClientState = iota
	Registered
)

func ClientThread(n *Node, ctx context.Context, done func()) {
	defer done()

	// Register with the orchestrator
	url := fmt.Sprintf("http://%s/register/%s", n.OrchAddr, n.Name)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Bad request")
		return
	}
	log.Println("Registered")

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
