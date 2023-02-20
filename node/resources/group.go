package resources

import (
	"context"

	"golang.org/x/sync/semaphore"
)

type ResourceGroup struct {
	ResourceBase
	semaphore semaphore.Weighted
	held      int64
}

func NewResourceGroup(name string, count int64) *ResourceGroup {
	return &ResourceGroup{
		ResourceBase: ResourceBase{
			Name:  name,
			Count: count,
		},
		semaphore: *semaphore.NewWeighted(count),
		held:      0,
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
	return r.held
}

func (r *ResourceGroup) Acquire(ctx context.Context, n int64) error {
	if err := r.semaphore.Acquire(ctx, n); err != nil {
		return err
	}
	r.held += n
	return nil
}

func (r *ResourceGroup) Release(n int64) {
	r.semaphore.Release(n)
	r.held -= n
	if r.held < 0 {
		r.held = 0
	}
}

func (r *ResourceGroup) TryAcquire(n int64) bool {
	ok := r.semaphore.TryAcquire(n)
	if ok {
		r.held += n
	}
	return ok
}
