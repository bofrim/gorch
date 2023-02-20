package resources

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

type ResourceBase struct {
	Name  string
	count int64
}

func (r *ResourceBase) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("Expected a map of names to counts, got %v", node.Kind)
	}
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		r.Name = keyNode.Value
		count, err := strconv.Atoi(valueNode.Value)
		if err != nil {
			return err
		}
		r.count = int64(count)
	}
	return nil
}

func (r *ResourceBase) UnmarshalJSON(data []byte) error {
	var m map[string]int64
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	for name, count := range m {
		r.Name = name
		r.count = count
	}
	return nil
}
