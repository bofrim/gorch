package node

import (
	"context"
	"log"
	"sync"
)

type Node struct {
	Name          string
	Port          int
	DataDir       string
	Data          map[string]map[string]interface{}
	ActionsPath   string
	Actions       map[string]Action
	OrchAddr      string
	OrchConnState ClientState
}

func (node *Node) Run() error {
	// Initialize Node
	// Load data
	data, err := loadData(node.DataDir)
	if err != nil {
		log.Fatal(err)
		return err
	}
	node.Data = data

	// Load actions
	if node.ActionsPath != "" {
		node.ReloadActions(node.ActionsPath)
	}

	// Run Node services
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	done := func() {
		wg.Done()
		cancel()
	}

	wg.Add(1)
	go MonitorThread(node, ctx, done)
	wg.Add(1)
	go ServerThread(node, ctx, done)

	// Finish
	wg.Wait()
	cancel()
	return nil
}

func (node *Node) ReloadActions(path string) error {
	// If no path was setup do nothing
	// Maybe pass some info back in the future
	if path == "" {
		return nil
	}
	actions, err := loadActions(node.ActionsPath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	node.Actions = actions
	node.ActionsPath = path
	return nil
}
