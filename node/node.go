package node

import (
	"context"
	"log"
	"sync"

	"github.com/bofrim/gorch/utils"
	"golang.org/x/exp/slog"
)

type Node struct {
	Name             string
	ServerAddr       string
	ServerPort       int
	DataDir          string
	Data             map[string]map[string]interface{}
	ActionsPath      string
	Actions          map[string]*Action
	OrchAddr         string
	nodeState        NodeState
	ArbitraryActions bool
	LogFile          string
}

func (node *Node) Run() (err error) {
	// Initialize Node
	var logger *slog.Logger
	var closeFn func()
	if logger, closeFn, err = utils.SetupLogging(node.LogFile); err != nil {
		return err
	}
	defer closeFn()

	// Load data
	data, err := loadData(node.DataDir)
	if err != nil {
		logger.Error("Failed to load data.", err)
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
	go MonitorThread(node, ctx, logger, done)
	wg.Add(1)
	go NServerThread(node, ctx, logger, done)
	wg.Add(1)
	go NodeStateThread(node, ctx, logger, done)

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
