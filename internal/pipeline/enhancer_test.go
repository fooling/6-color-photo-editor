package pipeline_test

import (
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

func TestNewEnhancerRegistry(t *testing.T) {
	r := pipeline.NewEnhancerRegistry()
	if r == nil {
		t.Fatal("NewEnhancerRegistry() returned nil")
	}
}

func TestEnhancerRegistry_RegisterAndGet(t *testing.T) {
	r := pipeline.NewEnhancerRegistry()

	// Create a basic enhancer
	e := pipeline.NewBasicEnhancer(0, 1, 1)
	r.Register(e)

	// Get it back
	got, ok := r.Get("basic")
	if !ok {
		t.Fatal("Get() returned false for registered enhancer")
	}
	if got.Name() != "basic" {
		t.Errorf("Expected name 'basic', got '%s'", got.Name())
	}
}

func TestEnhancerRegistry_GetUnregistered(t *testing.T) {
	r := pipeline.NewEnhancerRegistry()

	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get() returned true for unregistered enhancer")
	}
}

func TestEnhancerRegistry_List(t *testing.T) {
	r := pipeline.NewEnhancerRegistry()

	// Register multiple enhancers
	r.Register(pipeline.NewBasicEnhancer(0, 1, 1))
	r.Register(pipeline.NewAutoLevelsEnhancer(1.0))

	list := r.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 enhancers, got %d", len(list))
	}

	// Verify order is preserved
	if list[0].Name != "basic" {
		t.Errorf("Expected first enhancer to be 'basic', got '%s'", list[0].Name)
	}
	if list[1].Name != "auto_levels" {
		t.Errorf("Expected second enhancer to be 'auto_levels', got '%s'", list[1].Name)
	}
}

func TestEnhancerRegistry_RegisterDuplicate(t *testing.T) {
	r := pipeline.NewEnhancerRegistry()

	// Register the same enhancer twice
	r.Register(pipeline.NewBasicEnhancer(0, 1, 1))
	r.Register(pipeline.NewBasicEnhancer(0.5, 1.5, 1.5))

	// List should only have one entry
	list := r.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 enhancer after duplicate registration, got %d", len(list))
	}
}

func TestDefaultRegistry(t *testing.T) {
	// The default registry should have enhancers from init() calls
	list := pipeline.ListEnhancers()
	if len(list) == 0 {
		t.Error("Default registry is empty")
	}

	// Check that basic is registered
	_, ok := pipeline.GetEnhancer("basic")
	if !ok {
		t.Error("Basic enhancer not found in default registry")
	}
}

func TestEnhancerInfo(t *testing.T) {
	e := pipeline.NewBasicEnhancer(0, 1, 1)

	if e.Name() == "" {
		t.Error("Name() returned empty string")
	}
	if e.DisplayName() == "" {
		t.Error("DisplayName() returned empty string")
	}
	if e.Description() == "" {
		t.Error("Description() returned empty string")
	}
}
