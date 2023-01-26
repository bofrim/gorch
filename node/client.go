package node

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bofrim/gorch/orchestrator"
)

type ClientState int

const ClientPollPeriod = 5 * time.Second

const (
	Polling ClientState = iota
	Registered
	Disconnecting
	Disconnected
)

func ClientThread(n *Node, ctx context.Context, done func()) {
	defer done()
	n.OrchConnState = Polling
	ticker := time.NewTicker(ClientPollPeriod)
	for {
		select {
		case <-ticker.C:
			switch n.OrchConnState {
			case Polling:
				log.Println("Polling...")
				if err := register(n.OrchAddr, n.Name, n.ServerAddr, n.ServerPort); err == nil {
					n.OrchConnState = Registered
				}
			case Registered:
				log.Println("pinging")
				if err := ping(n.OrchAddr, n.Name); err != nil {
					log.Println("Bad ping. Going back to polling")
					n.OrchConnState = Polling
				}
			case Disconnecting:
				log.Println("Disconnecting.")
				_ = disconnect(n.OrchAddr, n.Name)
				n.OrchConnState = Disconnected
				return
			case Disconnected:
				log.Println("Disconnected.")
				return
			default:
				n.OrchConnState = Disconnecting
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
	log.Printf("Register with data: %s", b)
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
		return errors.New("request not OK")
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
	switch resp.StatusCode {
	case http.StatusOK:
		// success!
	default:
		// Something else
		fmt.Printf("Bad request: %s\n", resp.Status)
		return errors.New("request not OK")
	}

	log.Println("Pinged")
	return nil
}

func disconnect(addr, name string) error {

	return nil
}
