package node

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func simpleFileName(fpath string) string {
	fn := filepath.Base(fpath)
	ext := filepath.Ext(fn)
	return strings.TrimSuffix(fn, ext)
}

func loadData(baseDir string) (map[string]map[string]interface{}, error) {
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	return readFiles(baseDir, files)
}

func readFiles(baseDir string, files []fs.DirEntry) (map[string]map[string]interface{}, error) {
	var wg sync.WaitGroup
	data := make(map[string]map[string]interface{})
	errors := make([]error, len(files))

	for i := 0; i < len(files); i++ {
		file := files[i]
		index := i
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			fileData, err := readFile(filepath.Join(baseDir, file.Name()))
			if err != nil {
				errors[index] = err
				return
			}
			data[simpleFileName(file.Name())] = fileData
		}()
	}
	wg.Wait()

	for _, e := range errors {
		if e != nil {
			return nil, e
		}
	}
	return data, nil
}

func readFile(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var m map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
