// Package handler provides the HTTP handler and type registry for schema generation.
package handler

import (
	"fmt"
	"sync"
)

// Registry holds a mapping of type names to Go struct instances
// that can be used for schema generation by name.
type Registry struct {
	mu    sync.RWMutex
	types map[string]any
}

// NewRegistry creates an empty type registry.
func NewRegistry() *Registry {
	return &Registry{
		types: make(map[string]any),
	}
}

// Register adds a Go struct instance to the registry under the given name.
// If a type with the same name already exists, it will be overwritten.
func (r *Registry) Register(name string, v any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.types[name] = v
}

// Lookup returns the struct instance registered under the given name.
// Returns an error if the name is not found.
func (r *Registry) Lookup(name string) (any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.types[name]
	if !ok {
		return nil, fmt.Errorf("type %q not found in registry", name)
	}

	return v, nil
}

// Names returns a sorted list of all registered type names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.types))
	for name := range r.types {
		names = append(names, name)
	}

	return names
}
