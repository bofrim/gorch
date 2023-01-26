package orchestrator

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

func (nc *NodeConnection) RequestAction(actionName string, reqBody []byte) ([]byte, error) {
	log.Printf("Address: %s", nc.Address)
	log.Printf("Port: %d", nc.Port)
	log.Printf("Action: %s", actionName)
	url := fmt.Sprintf("http://%s:%d/action/%s", nc.Address, nc.Port, actionName)
	log.Printf("Sending action request to %s at address: %s", nc.Name, url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
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

	log.Printf("Finished action %s", actionName)
	log.Printf("The response was: %s", body)

	return body, nil

}

func (nc *NodeConnection) RequestData(reqBody []byte) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d/data", nc.Address, nc.Port)
	log.Printf("Sending action request to %s at address: %s", nc.Name, url)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(reqBody))
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("The response was: %s", respBody)
	return respBody, nil
}

func DisconnectThread(orchestrator *Orchestrator, ctx context.Context, done func()) {
	ticker := time.NewTicker(DisconnectionPeriod)
	for {
		select {
		case <-ticker.C:
			// Kick any nodes that we haven't heard from in the last DisconnectionPeriod
			for name, n := range orchestrator.Nodes {
				if n.LastInteraction.Before(time.Now().Add(-1 * DisconnectionPeriod)) {
					delete(orchestrator.Nodes, name)
					log.Printf("Kicked out node: %s\n", name)
				}
			}
		case <-ctx.Done():
			return
		}
	}

}
