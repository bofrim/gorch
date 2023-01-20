package node

import (
	"context"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func MonitorThread(node *Node, ctx context.Context, done func()) {
	defer done()

	// Create the watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return
	}
	defer watcher.Close()

	// Start watching
	go dataMonitor(*watcher, node, ctx)

	// Add the directory to be watched
	err = watcher.Add(node.DataDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Run until canceled
	<-ctx.Done()
}

func dataMonitor(watcher fsnotify.Watcher, node *Node, ctx context.Context) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if filepath.Ext(event.Name) == ".json" {
					fileData, err := readFile(event.Name)
					if err != nil {
						return
					}
					node.Data[simpleFileName(filepath.Base(event.Name))] = fileData
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Monitor error:", err)

		case <-ctx.Done():
			return

		}
	}
}
