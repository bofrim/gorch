package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

type ResourceManager struct {
	groups map[string]*ResourceGroup `yaml:"resource-groups`
}

func NewResourceManager(m map[string]int64) *ResourceManager {
	return &ResourceManager{
		groups: NewResourceGroupMap(m),
	}

}

func (rm *ResourceManager) GetCount(name string) int64 {
	return rm.groups[name].GetCount()
}

func (rm *ResourceManager) GetHeld(name string) int64 {
	return rm.groups[name].GetHeld()
}

func (rm *ResourceManager) Acquire(name string, ctx context.Context, n int64) error {
	return rm.groups[name].Acquire(ctx, n)
}

func (rm *ResourceManager) Release(name string, n int64) {
	rm.groups[name].Release(n)
}

func (rm *ResourceManager) TryAcquire(name string, n int64) bool {
	return rm.groups[name].TryAcquire(n)
}

func (r *ResourceManager) UnmarshalYAML(node *yaml.Node) error {
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

func (r *ResourceManager) UnmarshalJSON(data []byte) error {
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
