package pipeline_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

func TestNewEnhanceFilter(t *testing.T) {
	f := pipeline.NewEnhanceFilter(0.1, 1.2, 1.1)
	if f == nil {
		t.Fatal("NewEnhanceFilter() returned nil")
	}
}

func TestEnhanceFilter_Name(t *testing.T) {
	f := pipeline.NewEnhanceFilter(0.1, 1.2, 1.1)
	if f.Name() != "Enhance" {
		t.Errorf("Expected name 'Enhance', got '%s'", f.Name())
	}
}

func TestEnhanceFilter_Apply_Brightness(t *testing.T) {
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
			f := pipeline.NewEnhanceFilter(tt.bright, 1.0, 1.0)

			// Create a gray image
			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			gray := color.RGBA{128, 128, 128, 255}
			img.Set(5, 5, gray)

			result, err := f.Apply(img)
			if err != nil {
				t.Fatalf("Apply() returned error: %v", err)
			}

			if result == nil {
				t.Fatal("Apply() returned nil image")
			}
		})
	}
}

func TestEnhanceFilter_Apply_Contrast(t *testing.T) {
	tests := []struct {
		name    string
		contrast float64
	}{
		{"low contrast", 0.5},
		{"normal contrast", 1.0},
		{"high contrast", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := pipeline.NewEnhanceFilter(0, tt.contrast, 1.0)

			// Create a gray image
			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			gray := color.RGBA{128, 128, 128, 255}
			img.Set(5, 5, gray)

			result, err := f.Apply(img)
			if err != nil {
				t.Fatalf("Apply() returned error: %v", err)
			}

			if result == nil {
				t.Fatal("Apply() returned nil image")
			}
		})
	}
}

func TestEnhanceFilter_Apply_Saturation(t *testing.T) {
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
			f := pipeline.NewEnhanceFilter(0, 1.0, tt.saturation)

			// Create a red image
			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			red := color.RGBA{255, 0, 0, 255}
			img.Set(5, 5, red)

			result, err := f.Apply(img)
			if err != nil {
				t.Fatalf("Apply() returned error: %v", err)
			}

			if result == nil {
				t.Fatal("Apply() returned nil image")
			}
		})
	}
}

func TestEnhanceFilter_Apply_NoChange(t *testing.T) {
	// Filter with default values should still process
	f := pipeline.NewEnhanceFilter(0, 1.0, 1.0)

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{128, 128, 128, 255})

	result, err := f.Apply(img)
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

func TestEnhanceFilter_Apply_PreservesDimensions(t *testing.T) {
	f := pipeline.NewEnhanceFilter(0.2, 1.2, 1.1)

	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 50 {
		t.Errorf("Expected 100x50 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestEnhanceFilter_Apply_AllPixels(t *testing.T) {
	f := pipeline.NewEnhanceFilter(0.1, 1.0, 1.0)

	// Create a 3x3 image
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 3 || bounds.Dy() != 3 {
		t.Errorf("Expected 3x3 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Verify all pixels are accessible
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			c := result.At(x, y)
			if c == nil {
				t.Errorf("Pixel at (%d, %d) is nil", x, y)
			}
		}
	}
}

func TestEnhanceFilter_Apply_Clamping(t *testing.T) {
	// Test extreme values that would overflow
	tests := []struct {
		name       string
		brightness float64
	}{
		{"very negative brightness", -1.0},
		{"very positive brightness", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := pipeline.NewEnhanceFilter(tt.brightness, 1.0, 1.0)

			// Create a white image (testing overflow)
			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			white := color.RGBA{255, 255, 255, 255}
			img.Set(5, 5, white)

			result, err := f.Apply(img)
			if err != nil {
				t.Fatalf("Apply() returned error: %v", err)
			}

			// Get the resulting pixel
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

func TestEnhanceFilter_Apply_SaturationGrayscale(t *testing.T) {
	f := pipeline.NewEnhanceFilter(0, 1.0, 0) // saturation = 0 means grayscale

	// Create a color image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{255, 0, 0, 255}) // Red

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	// Get the resulting pixel
	c := result.At(5, 5)
	rgba, ok := c.(color.RGBA)
	if !ok {
		t.Fatal("Result is not RGBA")
	}

	// In grayscale, RGB should be equal (or very close)
	if rgba.R != rgba.G || rgba.G != rgba.B {
		// Allow some tolerance due to rounding
		diffRG := int(rgba.R) - int(rgba.G)
		diffGB := int(rgba.G) - int(rgba.B)
		if diffRG < -1 || diffRG > 1 || diffGB < -1 || diffGB > 1 {
			t.Errorf("Grayscale conversion failed: R=%d, G=%d, B=%d", rgba.R, rgba.G, rgba.B)
		}
	}
}
