package pipeline_test

import (
	"image"
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

// mockFilter implements pipeline.Filter for testing
type mockFilter struct {
	name        string
	applyFunc   func(img image.Image) (image.Image, error)
	applyCalled bool
}

func (m *mockFilter) Name() string {
	return m.name
}

func (m *mockFilter) Apply(img image.Image) (image.Image, error) {
	m.applyCalled = true
	if m.applyFunc != nil {
		return m.applyFunc(img)
	}
	return img, nil
}

func TestFilterInterface(t *testing.T) {
	// Test that mockFilter satisfies the Filter interface
	var _ pipeline.Filter = (*mockFilter)(nil)

	testImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	f := &mockFilter{name: "test"}
	if f.Name() != "test" {
		t.Errorf("Expected name 'test', got '%s'", f.Name())
	}

	_, err := f.Apply(testImg)
	if err != nil {
		t.Errorf("Apply() returned error: %v", err)
	}

	if !f.applyCalled {
		t.Error("Apply() was not called")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := pipeline.DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.Brightness != 0.0 {
		t.Errorf("Expected default brightness 0.0, got %v", config.Brightness)
	}

	if config.Contrast != 1.0 {
		t.Errorf("Expected default contrast 1.0, got %v", config.Contrast)
	}

	if config.Saturation != 1.0 {
		t.Errorf("Expected default saturation 1.0, got %v", config.Saturation)
	}

	if !config.Dither {
		t.Errorf("Expected default dither true, got %v", config.Dither)
	}
}

func TestProcessConfig_Defaults(t *testing.T) {
	config := &pipeline.ProcessConfig{}

	if config.Width != 0 {
		t.Errorf("Expected default width 0, got %d", config.Width)
	}

	if config.Height != 0 {
		t.Errorf("Expected default height 0, got %d", config.Height)
	}

	if config.Brightness != 0 {
		t.Errorf("Expected default brightness 0, got %v", config.Brightness)
	}

	if config.Contrast != 0 {
		t.Errorf("Expected default contrast 0, got %v", config.Contrast)
	}

	if config.Saturation != 0 {
		t.Errorf("Expected default saturation 0, got %v", config.Saturation)
	}

	if config.Dither {
		t.Errorf("Expected default dither false, got %v", config.Dither)
	}
}

func TestProcessingStep(t *testing.T) {
	step := pipeline.ProcessingStep{
		Name:  "test step",
		Image: nil,
	}

	if step.Name != "test step" {
		t.Errorf("Expected name 'test step', got '%s'", step.Name)
	}
}

func TestProcessingResult(t *testing.T) {
	result := &pipeline.ProcessingResult{
		Final:  nil,
		Steps:  nil,
		Config: nil,
	}

	if result.Final != nil {
		t.Error("Expected nil Final")
	}

	if result.Steps != nil {
		t.Error("Expected nil Steps")
	}

	if result.Config != nil {
		t.Error("Expected nil Config")
	}
}
