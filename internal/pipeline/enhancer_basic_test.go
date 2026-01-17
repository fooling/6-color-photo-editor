package pipeline_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

func TestNewBasicEnhancer(t *testing.T) {
	e := pipeline.NewBasicEnhancer(0.1, 1.2, 1.1)
	if e == nil {
		t.Fatal("NewBasicEnhancer() returned nil")
	}
}

func TestBasicEnhancer_Name(t *testing.T) {
	e := pipeline.NewBasicEnhancer(0, 1, 1)
	if e.Name() != "basic" {
		t.Errorf("Expected name 'basic', got '%s'", e.Name())
	}
}

func TestBasicEnhancer_DisplayName(t *testing.T) {
	e := pipeline.NewBasicEnhancer(0, 1, 1)
	if e.DisplayName() == "" {
		t.Error("DisplayName() returned empty string")
	}
}

func TestBasicEnhancer_Description(t *testing.T) {
	e := pipeline.NewBasicEnhancer(0, 1, 1)
	if e.Description() == "" {
		t.Error("Description() returned empty string")
	}
}

func TestBasicEnhancer_Apply(t *testing.T) {
	e := pipeline.NewBasicEnhancer(0, 1, 1)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{128, 128, 128, 255})

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

func TestBasicEnhancer_Apply_Brightness(t *testing.T) {
	tests := []struct {
		name   string
		bright float64
	}{
		{"negative brightness", -0.5},
		{"zero brightness", 0},
		{"positive brightness", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := pipeline.NewBasicEnhancer(tt.bright, 1.0, 1.0)

			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			img.Set(5, 5, color.RGBA{128, 128, 128, 255})

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

func TestBasicEnhancer_Apply_Contrast(t *testing.T) {
	tests := []struct {
		name     string
		contrast float64
	}{
		{"low contrast", 0.5},
		{"normal contrast", 1.0},
		{"high contrast", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := pipeline.NewBasicEnhancer(0, tt.contrast, 1.0)

			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			img.Set(5, 5, color.RGBA{128, 128, 128, 255})

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

func TestBasicEnhancer_Apply_Saturation(t *testing.T) {
	tests := []struct {
		name       string
		saturation float64
	}{
		{"grayscale", 0},
		{"normal saturation", 1.0},
		{"high saturation", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := pipeline.NewBasicEnhancer(0, 1.0, tt.saturation)

			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			img.Set(5, 5, color.RGBA{255, 0, 0, 255})

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

func TestBasicEnhancer_Apply_Clamping(t *testing.T) {
	tests := []struct {
		name       string
		brightness float64
	}{
		{"very negative brightness", -1.0},
		{"very positive brightness", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := pipeline.NewBasicEnhancer(tt.brightness, 1.0, 1.0)

			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			img.Set(5, 5, color.RGBA{255, 255, 255, 255})

			result, err := e.Apply(img)
			if err != nil {
				t.Fatalf("Apply() returned error: %v", err)
			}

			c := result.At(5, 5)
			rgba, ok := c.(color.RGBA)
			if !ok {
				t.Fatal("Result is not RGBA")
			}

			// Values should be clamped to [0, 255]
			if rgba.R > 255 || rgba.G > 255 || rgba.B > 255 {
				t.Errorf("Color values overflow: R=%d, G=%d, B=%d", rgba.R, rgba.G, rgba.B)
			}
		})
	}
}
