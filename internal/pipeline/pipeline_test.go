package pipeline_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

func TestNewPipeline(t *testing.T) {
	p := pipeline.NewPipeline(false)
	if p == nil {
		t.Fatal("NewPipeline() returned nil")
	}

	p2 := pipeline.NewPipeline(true)
	if p2 == nil {
		t.Fatal("NewPipeline(true) returned nil")
	}
}

func TestPipeline_AddFilter(t *testing.T) {
	p := pipeline.NewPipeline(false)

	f1 := &mockFilter{name: "filter1"}
	f2 := &mockFilter{name: "filter2"}

	// Test chaining
	result := p.AddFilter(f1).AddFilter(f2)

	if result != p {
		t.Error("AddFilter() should return the same pipeline for chaining")
	}
}

func TestPipeline_Clear(t *testing.T) {
	p := pipeline.NewPipeline(false)

	f1 := &mockFilter{name: "filter1"}
	p.AddFilter(f1)

	// Clear should remove all filters
	result := p.Clear()

	if result != p {
		t.Error("Clear() should return the same pipeline for chaining")
	}
}

func TestPipeline_Process_Empty(t *testing.T) {
	p := pipeline.NewPipeline(false)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	result := p.Process(img)

	if result == nil {
		t.Fatal("Process() returned nil result")
	}

	if result.Final == nil {
		t.Error("Process() returned nil final image")
	}

	if result.Config == nil {
		t.Error("Process() returned nil config")
	}
}

func TestPipeline_Process_SingleFilter(t *testing.T) {
	p := pipeline.NewPipeline(false)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	filterApplied := false
	f := &mockFilter{
		name: "test filter",
		applyFunc: func(img image.Image) (image.Image, error) {
			filterApplied = true
			return img, nil
		},
	}

	p.AddFilter(f)
	result := p.Process(img)

	if !filterApplied {
		t.Error("Filter was not applied")
	}

	if result.Final == nil {
		t.Error("Process() returned nil final image")
	}
}

func TestPipeline_Process_MultipleFilters(t *testing.T) {
	p := pipeline.NewPipeline(false)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	filtersApplied := make([]string, 0)
	f1 := &mockFilter{
		name: "filter1",
		applyFunc: func(img image.Image) (image.Image, error) {
			filtersApplied = append(filtersApplied, "filter1")
			return img, nil
		},
	}
	f2 := &mockFilter{
		name: "filter2",
		applyFunc: func(img image.Image) (image.Image, error) {
			filtersApplied = append(filtersApplied, "filter2")
			return img, nil
		},
	}
	f3 := &mockFilter{
		name: "filter3",
		applyFunc: func(img image.Image) (image.Image, error) {
			filtersApplied = append(filtersApplied, "filter3")
			return img, nil
		},
	}

	p.AddFilter(f1).AddFilter(f2).AddFilter(f3)
	result := p.Process(img)

	if len(filtersApplied) != 3 {
		t.Errorf("Expected 3 filters applied, got %d", len(filtersApplied))
	}

	if filtersApplied[0] != "filter1" || filtersApplied[1] != "filter2" || filtersApplied[2] != "filter3" {
		t.Errorf("Filters not applied in order: %v", filtersApplied)
	}

	if result.Final == nil {
		t.Error("Process() returned nil final image")
	}
}

func TestPipeline_Process_WithStepCapture(t *testing.T) {
	p := pipeline.NewPipeline(true)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	f1 := &mockFilter{name: "Resize"}
	f2 := &mockFilter{name: "Enhance"}
	f3 := &mockFilter{name: "Dither"}

	p.AddFilter(f1).AddFilter(f2).AddFilter(f3)
	result := p.Process(img)

	if result.Final == nil {
		t.Fatal("Process() returned nil final image")
	}

	// Should have captured steps for Resize, Enhance, and Dither
	if len(result.Steps) < 3 {
		t.Errorf("Expected at least 3 steps captured, got %d", len(result.Steps))
	}

	// Verify step names
	stepNames := make(map[string]bool)
	for _, step := range result.Steps {
		stepNames[step.Name] = true
	}

	if !stepNames["Resize"] {
		t.Error("Resize step not captured")
	}
	if !stepNames["Enhance"] {
		t.Error("Enhance step not captured")
	}
	if !stepNames["Dither"] {
		t.Error("Dither step not captured")
	}
}

func TestPipeline_Process_WithoutStepCapture(t *testing.T) {
	p := pipeline.NewPipeline(false)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	f1 := &mockFilter{name: "filter1"}
	f2 := &mockFilter{name: "filter2"}

	p.AddFilter(f1).AddFilter(f2)
	result := p.Process(img)

	if result.Final == nil {
		t.Fatal("Process() returned nil final image")
	}

	// Should not have captured steps when captureSteps is false
	if len(result.Steps) != 0 {
		t.Errorf("Expected 0 steps when captureSteps is false, got %d", len(result.Steps))
	}
}

func TestPipeline_Process_FilterError(t *testing.T) {
	p := pipeline.NewPipeline(false)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	f1 := &mockFilter{name: "filter1"}
	f2 := &mockFilter{
		name: "error filter",
		applyFunc: func(img image.Image) (image.Image, error) {
			return img, nil // No error for now
		},
	}
	f3 := &mockFilter{name: "filter3"}

	p.AddFilter(f1).AddFilter(f2).AddFilter(f3)
	result := p.Process(img)

	// Even without errors, process should complete
	if result.Final == nil {
		t.Error("Process() returned nil final image")
	}
}

func TestProcessWithConfig(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	config := &pipeline.ProcessConfig{
		Width:      50,
		Height:     50,
		Brightness: 0.1,
		Contrast:   1.2,
		Saturation: 1.1,
		Dither:     true,
	}

	result := pipeline.ProcessWithConfig(img, config, false)

	if result == nil {
		t.Fatal("ProcessWithConfig() returned nil result")
	}

	if result.Final == nil {
		t.Fatal("ProcessWithConfig() returned nil final image")
	}

	if result.Config != config {
		t.Error("ProcessWithConfig() did not set config in result")
	}

	// Verify dimensions changed
	bounds := result.Final.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Expected 50x50 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestProcessWithConfig_NoResize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	config := &pipeline.ProcessConfig{
		Width:      0,
		Height:     0,
		Brightness: 0.1,
		Contrast:   1.2,
		Saturation: 1.1,
		Dither:     true,
	}

	result := pipeline.ProcessWithConfig(img, config, false)

	if result == nil {
		t.Fatal("ProcessWithConfig() returned nil result")
	}

	if result.Final == nil {
		t.Fatal("ProcessWithConfig() returned nil final image")
	}

	// Dimensions should stay the same
	bounds := result.Final.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Expected 100x100 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestProcessWithConfig_OnlyDither(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	config := &pipeline.ProcessConfig{
		Dither: true,
	}

	result := pipeline.ProcessWithConfig(img, config, false)

	if result == nil {
		t.Fatal("ProcessWithConfig() returned nil result")
	}

	if result.Final == nil {
		t.Fatal("ProcessWithConfig() returned nil final image")
	}
}

func TestProcessWithConfig_WithStepCapture(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	config := &pipeline.ProcessConfig{
		Width:  5,
		Dither: true,
	}

	result := pipeline.ProcessWithConfig(img, config, true)

	if result == nil {
		t.Fatal("ProcessWithConfig() returned nil result")
	}

	if result.Final == nil {
		t.Fatal("ProcessWithConfig() returned nil final image")
	}

	// Should have captured steps
	if len(result.Steps) == 0 {
		t.Error("Expected steps to be captured with step capture enabled")
	}
}

func TestPipeline_Process_ChainedImageTransformation(t *testing.T) {
	p := pipeline.NewPipeline(false)

	// Create a test image with known pixel values
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{128, 128, 128, 255})

	// Add filters that actually transform the image
	resizeFilter := pipeline.NewResizeFilter(5, 5)
	enhanceFilter := pipeline.NewEnhanceFilter(0.1, 1.0, 1.0)

	p.AddFilter(resizeFilter).AddFilter(enhanceFilter)
	result := p.Process(img)

	if result.Final == nil {
		t.Fatal("Process() returned nil final image")
	}

	// Verify the image was resized
	bounds := result.Final.Bounds()
	if bounds.Dx() != 5 || bounds.Dy() != 5 {
		t.Errorf("Expected 5x5 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestPipeline_ChainingMethods(t *testing.T) {
	p := pipeline.NewPipeline(false)

	// Test that AddFilter returns the same pipeline
	f1 := &mockFilter{name: "filter1"}
	result := p.AddFilter(f1)

	if result != p {
		t.Error("AddFilter() should return the same pipeline")
	}

	// Test that Clear returns the same pipeline
	result = p.Clear()

	if result != p {
		t.Error("Clear() should return the same pipeline")
	}
}
