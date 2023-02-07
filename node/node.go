package node

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"golang.org/x/exp/slog"
	"golang.org/x/sync/semaphore"
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
	MaxNumActions    int
	CertPath         string
	actionSem        *semaphore.Weighted
}

func (node *Node) Run(logger *slog.Logger) (err error) {
	// Initialize Node
	node.actionSem = semaphore.NewWeighted(int64(node.MaxNumActions))
	logger.Debug("Created node semaphore.", slog.Int("count", node.MaxNumActions))

	// Load data
	if node.DataDir != "" {
		data, err := loadData(node.DataDir)
		if err != nil {
			logger.Error("Failed to load data.", err)
			return err
		}
		node.Data = data
	} else {
		node.Data = make(map[string]map[string]interface{})
	}

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

func (node *Node) RunAction(action *Action, streamDest string, params map[string]string, logger *slog.Logger) (out string, semOk bool, err error) {
	// First try to acquire the semaphore
	semOk = node.actionSem.TryAcquire(1)
	if !semOk {
		return out, semOk, err
	} else {
		// Next run the action
		// Ensure the semaphore is always released!
		if streamDest == "" {
			// Release asap
			defer node.actionSem.Release(1)
			outputs, err := action.Run(params)
			if err != nil {
				return out, semOk, err
			}
			out = strings.Join(outputs, "\n")
		} else {
			go func() {
				// Release when the action finishes
				defer node.actionSem.Release(1)
				action.RunStreamed(streamDest, params, logger)
			}()
			out = fmt.Sprintf("Streaming action output to %s", streamDest)
		}
	}

	return out, semOk, err
}
