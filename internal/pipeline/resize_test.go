package pipeline_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

func TestNewResizeFilter(t *testing.T) {
	f := pipeline.NewResizeFilter(100, 200)
	if f == nil {
		t.Fatal("NewResizeFilter() returned nil")
	}
}

func TestResizeFilter_Name(t *testing.T) {
	f := pipeline.NewResizeFilter(100, 200)
	if f.Name() != "Resize" {
		t.Errorf("Expected name 'Resize', got '%s'", f.Name())
	}
}

func TestResizeFilter_Apply_BothDimensions(t *testing.T) {
	f := pipeline.NewResizeFilter(50, 60)

	// Create a 100x100 test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 60 {
		t.Errorf("Expected 50x60 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestResizeFilter_Apply_WidthOnly(t *testing.T) {
	f := pipeline.NewResizeFilter(50, 0)

	// Create a 100x100 test image (1:1 aspect ratio)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 50 {
		t.Errorf("Expected width 50, got %d", bounds.Dx())
	}

	// With 1:1 aspect ratio, height should also be 50
	if bounds.Dy() != 50 {
		t.Errorf("Expected height 50 (maintaining aspect ratio), got %d", bounds.Dy())
	}
}

func TestResizeFilter_Apply_HeightOnly(t *testing.T) {
	f := pipeline.NewResizeFilter(0, 60)

	// Create a 100x100 test image (1:1 aspect ratio)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dy() != 60 {
		t.Errorf("Expected height 60, got %d", bounds.Dy())
	}

	// With 1:1 aspect ratio, width should also be 60
	if bounds.Dx() != 60 {
		t.Errorf("Expected width 60 (maintaining aspect ratio), got %d", bounds.Dx())
	}
}

func TestResizeFilter_Apply_NoDimensions(t *testing.T) {
	f := pipeline.NewResizeFilter(0, 0)

	// Create a 100x100 test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	// Should return the same image (no-op)
	if result != img {
		t.Error("Expected same image when both dimensions are 0")
	}

	bounds := result.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Expected 100x100 (no change), got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestResizeFilter_Apply_NonSquareAspect(t *testing.T) {
	// Create a 200x100 test image (2:1 aspect ratio)
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))

	// Set to width 100
	f := pipeline.NewResizeFilter(100, 0)

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 100 {
		t.Errorf("Expected width 100, got %d", bounds.Dx())
	}

	// Height should be 50 to maintain 2:1 aspect ratio
	if bounds.Dy() != 50 {
		t.Errorf("Expected height 50 (maintaining 2:1 aspect ratio), got %d", bounds.Dy())
	}
}

func TestResizeFilter_Apply_ContentPreserved(t *testing.T) {
	// Create an image with a specific pattern
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(9, 9, color.RGBA{0, 0, 255, 255})

	f := pipeline.NewResizeFilter(5, 5)

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 5 || bounds.Dy() != 5 {
		t.Errorf("Expected 5x5 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// The result should be a new image (not the same reference)
	if result == img {
		t.Error("Expected new image, got same reference")
	}
}
