package pipeline_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

func TestNewEInkOptimizedEnhancer(t *testing.T) {
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 0.3, 1.1)
	if e == nil {
		t.Fatal("NewEInkOptimizedEnhancer() returned nil")
	}
}

func TestEInkOptimizedEnhancer_Name(t *testing.T) {
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 0.3, 1.1)
	if e.Name() != "eink_optimized" {
		t.Errorf("Expected name 'eink_optimized', got '%s'", e.Name())
	}
}

func TestEInkOptimizedEnhancer_DisplayName(t *testing.T) {
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 0.3, 1.1)
	if e.DisplayName() == "" {
		t.Error("DisplayName() returned empty string")
	}
}

func TestEInkOptimizedEnhancer_Description(t *testing.T) {
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 0.3, 1.1)
	if e.Description() == "" {
		t.Error("Description() returned empty string")
	}
}

func TestEInkOptimizedEnhancer_Apply(t *testing.T) {
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 0.3, 1.1)

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

func TestEInkOptimizedEnhancer_Apply_NoSharpening(t *testing.T) {
	// Test with sharpening disabled
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 0.0, 1.1)

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

func TestEInkOptimizedEnhancer_Apply_NoColorBoost(t *testing.T) {
	// Test with color boost disabled
	e := pipeline.NewEInkOptimizedEnhancer(0.0, 0.3, 1.1)

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
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

func TestEInkOptimizedEnhancer_Apply_NoContrast(t *testing.T) {
	// Test with contrast at 1.0 (no change)
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 0.3, 1.0)

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

func TestEInkOptimizedEnhancer_Apply_PaletteColors(t *testing.T) {
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 0.3, 1.1)

	// Test with palette colors
	colors := []color.RGBA{
		{0, 0, 0, 255},       // Black
		{255, 255, 255, 255}, // White
		{255, 0, 0, 255},     // Red
		{0, 255, 0, 255},     // Green
		{0, 0, 255, 255},     // Blue
		{255, 255, 0, 255},   // Yellow
	}

	for _, c := range colors {
		img := image.NewRGBA(image.Rect(0, 0, 5, 5))
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				img.Set(x, y, c)
			}
		}

		result, err := e.Apply(img)
		if err != nil {
			t.Fatalf("Apply() returned error for color %v: %v", c, err)
		}
		if result == nil {
			t.Fatalf("Apply() returned nil image for color %v", c)
		}
	}
}

func TestEInkOptimizedEnhancer_Apply_EdgeHandling(t *testing.T) {
	// Test that edge pixels are handled correctly during sharpening
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 1.0, 1.1)

	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
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

	// Check all corner pixels are valid
	corners := []image.Point{{0, 0}, {2, 0}, {0, 2}, {2, 2}}
	for _, p := range corners {
		c := result.At(p.X, p.Y)
		if c == nil {
			t.Errorf("Corner pixel at (%d, %d) is nil", p.X, p.Y)
		}
	}
}

func TestEInkOptimizedEnhancer_Apply_SinglePixel(t *testing.T) {
	e := pipeline.NewEInkOptimizedEnhancer(0.5, 0.5, 1.1)

	// Test with a 1x1 image
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{128, 128, 128, 255})

	result, err := e.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Apply() returned nil image")
	}
}

func TestEInk6ColorPalette(t *testing.T) {
	// Verify the palette has 6 colors
	if len(pipeline.EInk6ColorPalette) != 6 {
		t.Errorf("Expected 6 colors in palette, got %d", len(pipeline.EInk6ColorPalette))
	}
}
