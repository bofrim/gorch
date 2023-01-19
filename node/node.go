package node

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Node struct {
	Port    int
	DataDir string
	Data    map[string]interface{}
}

// The first version of a node will be responsible for 2 things
// 1. Loading and watching data from the directory
// 2. Serving a simple rest endpoint that can be used to read the data
func (n *Node) Run() error {
	log.Println("Starting Node")
	// First load all data
	data, err := ReadDataDir(n.DataDir)
	if err != nil {
		return err
	}
	n.Data = data

	// Next start the watcher thread
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	errChan := make(chan error)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println(event)
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					if filepath.Ext(event.Name) == ".json" {
						fileData, err := LoadFile(event.Name)
						if err != nil {
							errChan <- err
							return
						}
						n.Data[simpleFileName(filepath.Base(event.Name))] = fileData
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add(n.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
	return nil
}

func ReadDataDir(baseDir string) (map[string]interface{}, error) {
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		fileData, err := LoadFile(filepath.Join(baseDir, file.Name()))
		if err != nil {
			return nil, err
		}
		data[simpleFileName(file.Name())] = fileData
	}
	return data, nil
}

func LoadFile(filePath string) (interface{}, error) {
	var err error
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var m interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&m)
	if err != nil {
		return nil, err
	}

	log.Printf("%s: %+v", simpleFileName(filePath), m)

	return m, nil
}

func simpleFileName(fpath string) string {
	fn := filepath.Base(fpath)
	ext := filepath.Ext(fn)
	return strings.TrimSuffix(fn, ext)
}
