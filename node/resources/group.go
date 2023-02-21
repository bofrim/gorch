package resources

import (
	"context"

	"golang.org/x/sync/semaphore"
)

type ResourceGroup struct {
	ResourceBase
	semaphore semaphore.Weighted
	Held      int64 `json:"held"`
}

func NewResourceGroup(name string, count int64) *ResourceGroup {
	return &ResourceGroup{
		ResourceBase: ResourceBase{
			Name:  name,
			Count: count,
		},
		semaphore: *semaphore.NewWeighted(count),
		Held:      0,
	}
}

func NewResourceGroupMap(m map[string]int64) map[string]*ResourceGroup {
	groups := map[string]*ResourceGroup{}
	for name, count := range m {
		groups[name] = NewResourceGroup(name, count)
	}
	return groups
}

func (r *ResourceGroup) GetCount() int64 {
	return r.Count
}

func (r *ResourceGroup) GetHeld() int64 {
	return r.Held
}

func (r *ResourceGroup) Acquire(ctx context.Context, n int64) error {
	if err := r.semaphore.Acquire(ctx, n); err != nil {
		return err
	}
	r.Held += n
	return nil
}

func (r *ResourceGroup) Release(n int64) {
	r.semaphore.Release(n)
	r.Held -= n
	if r.Held < 0 {
		r.Held = 0
	}
}

func (r *ResourceGroup) TryAcquire(n int64) bool {
	ok := r.semaphore.TryAcquire(n)
	if ok {
		r.Held += n
	}
	return ok
}
