package pipeline_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

func TestNewAutoLevelsEnhancer(t *testing.T) {
	e := pipeline.NewAutoLevelsEnhancer(1.0)
	if e == nil {
		t.Fatal("NewAutoLevelsEnhancer() returned nil")
	}
}

func TestAutoLevelsEnhancer_Name(t *testing.T) {
	e := pipeline.NewAutoLevelsEnhancer(1.0)
	if e.Name() != "auto_levels" {
		t.Errorf("Expected name 'auto_levels', got '%s'", e.Name())
	}
}

func TestAutoLevelsEnhancer_DisplayName(t *testing.T) {
	e := pipeline.NewAutoLevelsEnhancer(1.0)
	if e.DisplayName() == "" {
		t.Error("DisplayName() returned empty string")
	}
}

func TestAutoLevelsEnhancer_Description(t *testing.T) {
	e := pipeline.NewAutoLevelsEnhancer(1.0)
	if e.Description() == "" {
		t.Error("Description() returned empty string")
	}
}

func TestAutoLevelsEnhancer_Apply(t *testing.T) {
	e := pipeline.NewAutoLevelsEnhancer(1.0)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 25), uint8(y * 25), 128, 255})
		}
	}

	result, err := e.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Apply() returned nil image")
	}

	bounds := result.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("Expected 10x10 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestAutoLevelsEnhancer_Apply_LowContrast(t *testing.T) {
	e := pipeline.NewAutoLevelsEnhancer(0.5)

	// Create a low-contrast image (all pixels in the 100-150 range)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			val := uint8(100 + (x+y)%50)
			img.Set(x, y, color.RGBA{val, val, val, 255})
		}
	}

	result, err := e.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Apply() returned nil image")
	}
}

func TestAutoLevelsEnhancer_Apply_UniformImage(t *testing.T) {
	e := pipeline.NewAutoLevelsEnhancer(1.0)

	// Create a uniform image (all same color)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}

	result, err := e.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Apply() returned nil image")
	}
}

func TestAutoLevelsEnhancer_Apply_DifferentClipPercent(t *testing.T) {
	tests := []struct {
		name        string
		clipPercent float64
	}{
		{"no clipping", 0.0},
		{"1% clipping", 1.0},
		{"5% clipping", 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := pipeline.NewAutoLevelsEnhancer(tt.clipPercent)

			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			for y := 0; y < 10; y++ {
				for x := 0; x < 10; x++ {
					img.Set(x, y, color.RGBA{uint8(x * 25), uint8(y * 25), 128, 255})
				}
			}

			result, err := e.Apply(img)
			if err != nil {
				t.Fatalf("Apply() returned error: %v", err)
			}
			if result == nil {
				t.Fatal("Apply() returned nil image")
			}
		})
	}
}

func TestAutoLevelsEnhancer_Apply_FullRange(t *testing.T) {
	e := pipeline.NewAutoLevelsEnhancer(0.0)

	// Create an image with full dynamic range
	img := image.NewRGBA(image.Rect(0, 0, 256, 1))
	for x := 0; x < 256; x++ {
		img.Set(x, 0, color.RGBA{uint8(x), uint8(x), uint8(x), 255})
	}

	result, err := e.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Apply() returned nil image")
	}
}
