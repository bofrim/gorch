package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type ResourceHandle struct {
	id      uuid.UUID
	Request *ResourceRequest `json:"request"`
	Created time.Time        `json:"created"`
}

type ResourceManager struct {
	Groups map[string]*ResourceGroup     `json:"groups"`
	Active map[uuid.UUID]*ResourceHandle `json:"active"`
}

func NewResourceManager(m map[string]int64) *ResourceManager {
	return &ResourceManager{
		Groups: NewResourceGroupMap(m),
		Active: map[uuid.UUID]*ResourceHandle{},
	}
}

func (rm *ResourceManager) GetCount(name string) int64 {
	return rm.Groups[name].GetCount()
}

func (rm *ResourceManager) GetHeld(name string) int64 {
	return rm.Groups[name].GetHeld()
}

func (rm *ResourceManager) Acquire(name string, ctx context.Context, n int64) error {
	return rm.Groups[name].Acquire(ctx, n)
}

func (rm *ResourceManager) TryAcquireRequest(request *ResourceRequest) (uuid.UUID, error) {
	acquired := []*ResourceBase{}
	success := true
	for _, resource := range request.Resources {
		if ok := rm.TryAcquire(resource.Name, resource.Count); ok {
			acquired = append(acquired, resource)
		} else {
			success = false
			break
		}
	}
	if success {
		handle := ResourceHandle{
			id:      uuid.Must(uuid.NewUUID()),
			Request: request,
			Created: time.Now(),
		}
		rm.Active[handle.id] = &handle
		return handle.id, nil
	} else {
		// If we didn't succeed, put everything back
		for _, resource := range acquired {
			rm.Release(resource.Name, resource.Count)
		}
		return uuid.UUID{}, fmt.Errorf("unable to acquire resources")
	}
}

func (rm *ResourceManager) TryAcquire(name string, n int64) bool {
	return rm.Groups[name].TryAcquire(n)
}

func (rm *ResourceManager) ReleaseHandle(id uuid.UUID) {
	// Pop the handle out of the map
	handle, ok := rm.Active[id]
	if !ok {
		return
	} else {
		delete(rm.Active, id)
	}
	// Release all held resources
	for _, resource := range handle.Request.Resources {
		rm.Release(resource.Name, resource.Count)
	}
}

func (rm *ResourceManager) Release(name string, n int64) {
	rm.Groups[name].Release(n)
}

func (r *ResourceManager) UnmarshalYAML(node *yaml.Node) error {
	var m map[string]int64
	if err := node.Decode(m); err != nil {
		return err
	}

	r.Groups = NewResourceGroupMap(m)
	return nil
}

func (r *ResourceManager) UnmarshalJSON(data []byte) error {
	var m map[string]int64
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	r.Groups = NewResourceGroupMap(m)
	return nil
}
