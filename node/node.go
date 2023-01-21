package node

import (
	"context"
	"log"
	"sync"
)

type Node struct {
	Port        int
	DataDir     string
	Data        map[string]map[string]interface{}
	ActionsPath string
	Actions     map[string]Action
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
	actions, err := loadActions(node.ActionsPath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	node.Actions = actions

	// Run Node services
	var wg sync.WaitGroup
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())

	go MonitorThread(node, ctx, wg.Done)
	go ServerThread(node, ctx, wg.Done)

	// Finish
	wg.Wait()
	cancel()
	return nil
}
