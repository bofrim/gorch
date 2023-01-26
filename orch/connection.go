package orch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const DisconnectionPeriod = 10 * time.Second

type NodeConnection struct {
	Name            string
	Address         string
	Port            int
	LastInteraction time.Time
}

func (nc *NodeConnection) RequestAction(name string, data []byte) ([]byte, error) {
	log.Printf("Address: %s", nc.Address)
	log.Printf("Port: %d", nc.Port)
	log.Printf("Action: %s", name)
	url := fmt.Sprintf("http://%s:%d/action/%s", nc.Address, nc.Port, name)
	log.Printf("Sending action request to %s at address: %s", nc.Name, url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		// success!
	default:
		// Something else
		fmt.Printf("Bad request: %s\n", resp.Status)
		return nil, errors.New("request not OK")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Finished action %s", name)
	log.Printf("The response was: %s", body)

	return body, nil

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
