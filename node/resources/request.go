package resources

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type ResourceRequest struct {
	Resources map[string]*ResourceBase
}

func (r *ResourceRequest) UnmarshalYAML(node *yaml.Node) error {
	var m map[string]int64
	if err := node.Decode(&m); err != nil {
		return err
	}

	r.Resources = NewResourceBaseMap(m)
	return nil
}

func (r *ResourceRequest) UnmarshalJSON(data []byte) error {
	var m map[string]int64
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	r.Resources = NewResourceBaseMap(m)
	return nil
}
