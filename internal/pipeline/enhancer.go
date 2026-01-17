package pipeline

import (
	"image"
	"sync"
)

// Enhancer defines the interface for image enhancement algorithms.
// Implementations can provide different enhancement strategies optimized
// for various use cases (e.g., general purpose, E-Ink displays, etc.)
type Enhancer interface {
	// Name returns a unique identifier for the enhancer (used in API)
	Name() string

	// DisplayName returns a human-readable name for UI display
	DisplayName() string

	// Description returns a brief description of what this enhancer does
	Description() string

	// Apply applies the enhancement to an image
	Apply(img image.Image) (image.Image, error)
}

// EnhancerInfo contains metadata about an enhancer for API responses
type EnhancerInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// EnhancerRegistry manages available enhancers
type EnhancerRegistry struct {
	mu        sync.RWMutex
	enhancers map[string]Enhancer
	order     []string // maintains registration order
}

// NewEnhancerRegistry creates a new enhancer registry
func NewEnhancerRegistry() *EnhancerRegistry {
	return &EnhancerRegistry{
		enhancers: make(map[string]Enhancer),
		order:     make([]string, 0),
	}
}

// Register adds an enhancer to the registry
func (r *EnhancerRegistry) Register(e Enhancer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := e.Name()
	if _, exists := r.enhancers[name]; !exists {
		r.order = append(r.order, name)
	}
	r.enhancers[name] = e
}

// Get retrieves an enhancer by name
func (r *EnhancerRegistry) Get(name string) (Enhancer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.enhancers[name]
	return e, ok
}

// List returns information about all registered enhancers in registration order
func (r *EnhancerRegistry) List() []EnhancerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]EnhancerInfo, 0, len(r.order))
	for _, name := range r.order {
		if e, ok := r.enhancers[name]; ok {
			result = append(result, EnhancerInfo{
				Name:        e.Name(),
				DisplayName: e.DisplayName(),
				Description: e.Description(),
			})
		}
	}
	return result
}

// DefaultRegistry is the global enhancer registry
var DefaultRegistry = NewEnhancerRegistry()

// RegisterEnhancer registers an enhancer to the default registry
func RegisterEnhancer(e Enhancer) {
	DefaultRegistry.Register(e)
}

// GetEnhancer retrieves an enhancer from the default registry
func GetEnhancer(name string) (Enhancer, bool) {
	return DefaultRegistry.Get(name)
}

// ListEnhancers returns all enhancers from the default registry
func ListEnhancers() []EnhancerInfo {
	return DefaultRegistry.List()
}
