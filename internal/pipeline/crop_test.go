package pipeline_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/pipeline"
)

func TestNewCropFilter(t *testing.T) {
	f := pipeline.NewCropFilter(0.25, 0.25, 0.5, 0.5)
	if f == nil {
		t.Fatal("NewCropFilter() returned nil")
	}
}

func TestCropFilter_Name(t *testing.T) {
	f := pipeline.NewCropFilter(0.25, 0.25, 0.5, 0.5)
	if f.Name() != "Crop" {
		t.Errorf("Expected name 'Crop', got '%s'", f.Name())
	}
}

func TestCropFilter_Apply_CenterCrop(t *testing.T) {
	f := pipeline.NewCropFilter(0.25, 0.25, 0.5, 0.5)

	// Create a 100x100 test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill with different colors in each quadrant
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			var c color.RGBA
			if x < 50 && y < 50 {
				c = color.RGBA{255, 0, 0, 255} // Top-left: red
			} else if x >= 50 && y < 50 {
				c = color.RGBA{0, 255, 0, 255} // Top-right: green
			} else if x < 50 && y >= 50 {
				c = color.RGBA{0, 0, 255, 255} // Bottom-left: blue
			} else {
				c = color.RGBA{255, 255, 0, 255} // Bottom-right: yellow
			}
			img.Set(x, y, c)
		}
	}

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	// Should crop to center 50x50 (from x=25,y=25 to x=75,y=75)
	bounds := result.Bounds()

	// Check bounds start at (0, 0)
	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		t.Errorf("Expected bounds to start at (0, 0), got (%d, %d)", bounds.Min.X, bounds.Min.Y)
	}

	// Check dimensions
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Expected 50x50 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Check that we can access pixels at (0,0)
	c := result.At(0, 0)
	if c == nil {
		t.Error("Cannot access pixel at (0, 0)")
	}

	// Check a pixel in the result - should be from center area
	// At (0, 0) in result corresponds to (25, 25) in original
	// which should be red (from top-left quadrant)
	r, g, b, _ := result.At(0, 0).RGBA()
	if r>>8 < 200 || g>>8 > 50 || b>>8 > 50 {
		t.Errorf("Expected red pixel at (0,0), got R=%d G=%d B=%d", r>>8, g>>8, b>>8)
	}
}

func TestCropFilter_Apply_OffsetCrop(t *testing.T) {
	// Crop from (0.2, 0.3) with size (0.4, 0.5)
	f := pipeline.NewCropFilter(0.2, 0.3, 0.4, 0.5)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 128, 255})
		}
	}

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()

	// Check bounds start at (0, 0)
	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		t.Errorf("Expected bounds to start at (0, 0), got (%d, %d)", bounds.Min.X, bounds.Min.Y)
	}

	// Should be 40x50 (0.4*100 x 0.5*100)
	if bounds.Dx() != 40 || bounds.Dy() != 50 {
		t.Errorf("Expected 40x50 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Pixel at (0, 0) in result corresponds to (20, 30) in original
	r, g, _, _ := result.At(0, 0).RGBA()
	expectedR := uint8(20)
	expectedG := uint8(30)
	if r>>8 != uint32(expectedR) || g>>8 != uint32(expectedG) {
		t.Errorf("Expected pixel (%d, %d) at (0,0), got (%d, %d)", expectedR, expectedG, r>>8, g>>8)
	}
}

func TestCropFilter_Apply_FullImage(t *testing.T) {
	// Crop entire image (0, 0, 1.0, 1.0)
	f := pipeline.NewCropFilter(0, 0, 1.0, 1.0)

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Expected 50x50 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestCropFilter_Apply_Clamping(t *testing.T) {
	// Request crop beyond image bounds
	f := pipeline.NewCropFilter(0.8, 0.8, 0.5, 0.5)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	// Should be clamped to available space (20x20)
	bounds := result.Bounds()
	if bounds.Dx() != 20 || bounds.Dy() != 20 {
		t.Errorf("Expected 20x20 result (clamped), got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestNewCropToAspectFilter(t *testing.T) {
	f := pipeline.NewCropToAspectFilter(16, 9)
	if f == nil {
		t.Fatal("NewCropToAspectFilter() returned nil")
	}
}

func TestCropToAspectFilter_Name(t *testing.T) {
	f := pipeline.NewCropToAspectFilter(16, 9)
	if f.Name() != "CropAspect" {
		t.Errorf("Expected name 'CropAspect', got '%s'", f.Name())
	}
}

func TestCropToAspectFilter_Apply_WiderImage(t *testing.T) {
	// Target: 16:9, Image: 200x100 (2:1, wider than target)
	f := pipeline.NewCropToAspectFilter(16, 9)

	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()

	// Check bounds start at (0, 0)
	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		t.Errorf("Expected bounds to start at (0, 0), got (%d, %d)", bounds.Min.X, bounds.Min.Y)
	}

	// Should crop width to match aspect ratio
	// height stays 100, width becomes 100 * 16/9 ≈ 177
	expectedWidth := 177
	if bounds.Dx() != expectedWidth {
		t.Errorf("Expected width %d, got %d", expectedWidth, bounds.Dx())
	}
	if bounds.Dy() != 100 {
		t.Errorf("Expected height 100, got %d", bounds.Dy())
	}
}

func TestCropToAspectFilter_Apply_TallerImage(t *testing.T) {
	// Target: 16:9, Image: 100x200 (1:2, taller than target)
	f := pipeline.NewCropToAspectFilter(16, 9)

	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()

	// Check bounds start at (0, 0)
	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		t.Errorf("Expected bounds to start at (0, 0), got (%d, %d)", bounds.Min.X, bounds.Min.Y)
	}

	// Should crop height to match aspect ratio
	// width stays 100, height becomes 100 / (16/9) = 56.25 ≈ 56
	expectedHeight := 56
	if bounds.Dy() != expectedHeight {
		t.Errorf("Expected height %d, got %d", expectedHeight, bounds.Dy())
	}
	if bounds.Dx() != 100 {
		t.Errorf("Expected width 100, got %d", bounds.Dx())
	}
}

func TestCropToAspectFilter_Apply_Centering(t *testing.T) {
	f := pipeline.NewCropToAspectFilter(1, 1)

	// Create 100x50 image (wide)
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))

	// Fill left and right with different colors
	for y := 0; y < 50; y++ {
		for x := 0; x < 100; x++ {
			if x < 25 {
				img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Left: red
			} else if x >= 75 {
				img.Set(x, y, color.RGBA{0, 0, 255, 255}) // Right: blue
			} else {
				img.Set(x, y, color.RGBA{0, 255, 0, 255}) // Center: green
			}
		}
	}

	result, err := f.Apply(img)
	if err != nil {
		t.Fatalf("Apply() returned error: %v", err)
	}

	bounds := result.Bounds()

	// Should be 50x50 (cropped to square, centered)
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("Expected 50x50 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Center pixel should be green (from center area)
	_, g, _, _ := result.At(25, 25).RGBA()
	if g>>8 < 200 {
		t.Errorf("Expected green pixel in center, got G=%d", g>>8)
	}

	// Should not have red or blue (cropped from sides)
	r, _, b, _ := result.At(0, 0).RGBA()
	if r>>8 > 50 || b>>8 > 50 {
		t.Errorf("Expected no red/blue after centering, got R=%d B=%d", r>>8, b>>8)
	}
}
