package provider

import (
	"fmt"
	"sort"
	"sync"
)

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[string]Adapter)}
}

func (r *Registry) Register(a Adapter) error {
	if a == nil || a.ID() == "" {
		return fmt.Errorf("provider adapter requires a non-empty identity")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.adapters[a.ID()]; exists {
		return fmt.Errorf("provider %q already registered", a.ID())
	}
	r.adapters[a.ID()] = a
	return nil
}

func (r *Registry) Get(id string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[id]
	return a, ok
}

func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.adapters))
	for id := range r.adapters {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
