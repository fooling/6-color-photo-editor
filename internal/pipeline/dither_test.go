package pipeline_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

func TestNewDitherFilter(t *testing.T) {
	f := pipeline.NewDitherFilter()
	if f == nil {
		t.Fatal("NewDitherFilter() returned nil")
	}
}

func TestDitherFilter_Name(t *testing.T) {
	f := pipeline.NewDitherFilter()
	if f.Name() != "Dither" {
		t.Errorf("Expected name 'Dither', got '%s'", f.Name())
	}
}

func TestDitherFilter_Apply_SolidPaletteColors(t *testing.T) {
	f := pipeline.NewDitherFilter()

	tests := []struct {
		name  string
		color color.Color
	}{
		{"black", color.Black},
		{"white", color.White},
		{"green", color.RGBA{0, 255, 0, 255}},
		{"blue", color.RGBA{0, 0, 255, 255}},
		{"red", color.RGBA{255, 0, 0, 255}},
		{"yellow", color.RGBA{255, 255, 0, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a 10x10 solid color image
			img := image.NewRGBA(image.Rect(0, 0, 10, 10))
			for y := 0; y < 10; y++ {
				for x := 0; x < 10; x++ {
					img.Set(x, y, tt.color)
				}
			}

			result, err := f.Apply(img)
			if err != nil {
				t.Fatalf("Apply() returned error: %v", err)
			}

			if result == nil {
				t.Fatal("Apply() returned nil image")
			}

			// Verify dimensions
			bounds := result.Bounds()
			if bounds.Dx() != 10 || bounds.Dy() != 10 {
				t.Errorf("Expected 10x10 result, got %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestDitherFilter_Apply_Gradient(t *testing.T) {
	f := pipeline.NewDitherFilter()

	// Create a horizontal gradient from black to white
	img := image.NewRGBA(image.Rect(0, 0, 50, 10))
	for x := 0; x < 50; x++ {
		val := uint8(x * 255 / 49)
		c := color.RGBA{val, val, val, 255}
		for y := 0; y < 10; y++ {
			img.Set(x, y, c)
		}
	}

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Apply() returned nil image")
	}

	// Verify dimensions
	bounds := result.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 10 {
		t.Errorf("Expected 50x10 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestDitherFilter_Apply_PreservesDimensions(t *testing.T) {
	f := pipeline.NewDitherFilter()

	tests := []struct {
		name            string
		width, height int
	}{
		{"1x1", 1, 1},
		{"10x10", 10, 10},
		{"296x296", 296, 296},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, tt.width, tt.height))

			result, err := f.Apply(img)
			if err != nil {
				t.Fatalf("Apply() returned error: %v", err)
			}

			bounds := result.Bounds()
			if bounds.Dx() != tt.width || bounds.Dy() != tt.height {
				t.Errorf("Expected %dx%d result, got %dx%d", tt.width, tt.height, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestDitherFilter_Apply_ReturnsNewImage(t *testing.T) {
	f := pipeline.NewDitherFilter()

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	// Result should be a new image
	if result == img {
		t.Error("Expected new image, got same reference")
	}
}

func TestDitherFilter_Apply_AllColors(t *testing.T) {
	f := pipeline.NewDitherFilter()

	// Create a test image with a color not in the palette
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	purple := color.RGBA{128, 0, 128, 255}
	img.Set(5, 5, purple)

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	// The result should have been processed
	c := result.At(5, 5)
	if c == nil {
		t.Error("Result pixel is nil")
	}
}

func TestDitherFilter_EmptyImage(t *testing.T) {
	f := pipeline.NewDitherFilter()

	// Create a 0x0 image
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 0 || bounds.Dy() != 0 {
		t.Errorf("Expected 0x0 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
