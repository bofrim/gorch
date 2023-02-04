package node

import (
	"context"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/exp/slog"
)

func MonitorThread(node *Node, ctx context.Context, logger *slog.Logger, done func()) {
	defer done()

	// Create the watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("Failed to create a new file watcher.", err)
		return
	}
	defer watcher.Close()

	// Start watching
	go dataMonitor(*watcher, node, ctx, logger, done)

	// Add the directory to be watched
	err = watcher.Add(node.DataDir)
	if err != nil {
		logger.Error("Failed to add the data directory to the watcher.", err)
		return
	}

	// Run until canceled
	<-ctx.Done()
}

func dataMonitor(watcher fsnotify.Watcher, node *Node, ctx context.Context, logger *slog.Logger, done func()) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				logger.Warn("Watcher got a not ok event.", slog.Any("event", event))
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				logger.Debug("Watcher got an event.", slog.Any("event", event))
				if filepath.Ext(event.Name) == ".json" {
					fileData, err := readFile(event.Name)
					if err != nil {
						logger.Warn("Watcher got a not ok event.", slog.Any("event", event))
						return
					}
					node.Data[simpleFileName(filepath.Base(event.Name))] = fileData
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				logger.Warn("Watcher error not OK.")
				return
			}
			logger.Error("Watcher error.", err)

		case <-ctx.Done():
			logger.Info("Watcher done.")
			return
		}
	}
}
