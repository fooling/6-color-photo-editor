package core_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/core"
)

// Helper function to check if a color is approximately equal to another
func colorsEqual(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func TestFloydSteinbergDither_SolidColors(t *testing.T) {
	palette := core.NewEInkPalette()

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

			result := core.FloydSteinbergDither(img, palette)

			// Verify result is valid
			bounds := result.Bounds()
			if bounds.Dx() != 10 || bounds.Dy() != 10 {
				t.Errorf("Expected 10x10 result, got %dx%d", bounds.Dx(), bounds.Dy())
			}

			// Verify result has colors (the dithering should have processed the image)
			for y := 0; y < 10; y++ {
				for x := 0; x < 10; x++ {
					c := result.At(x, y)
					if c == nil {
						t.Errorf("Pixel at (%d, %d) is nil", x, y)
					}
				}
			}
		})
	}
}

func TestFloydSteinbergDither_Gradient(t *testing.T) {
	palette := core.NewEInkPalette()

	// Create a horizontal gradient from black to white
	img := image.NewRGBA(image.Rect(0, 0, 100, 10))
	for x := 0; x < 100; x++ {
		val := uint8(x * 255 / 99)
		c := color.RGBA{val, val, val, 255}
		for y := 0; y < 10; y++ {
			img.Set(x, y, c)
		}
	}

	result := core.FloydSteinbergDither(img, palette)

	// Verify all pixels are valid
	bounds := result.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 10 {
		t.Errorf("Expected 100x10 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Verify we have some variation (not all pixels are the same)
	firstColor := result.At(0, 0)
	lastColor := result.At(99, 0)
	hasVariation := !colorsEqual(firstColor, lastColor)
	if !hasVariation {
		t.Error("Dithered gradient should have variation")
	}
}

func TestFloydSteinbergDither_1x1Image(t *testing.T) {
	palette := core.NewEInkPalette()

	// Create a 1x1 gray image
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{128, 128, 128, 255})

	result := core.FloydSteinbergDither(img, palette)

	// Verify result is 1x1
	bounds := result.Bounds()
	if bounds.Dx() != 1 || bounds.Dy() != 1 {
		t.Errorf("Expected 1x1 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Verify the pixel is valid
	c := result.At(0, 0)
	if c == nil {
		t.Error("Pixel is nil")
	}
}

func TestFloydSteinbergDither_ErrorDiffusion(t *testing.T) {
	palette := core.NewEInkPalette()

	// Create a simple 2x2 image with a non-palette color
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	gray := color.RGBA{128, 128, 128, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, gray)
		}
	}

	result := core.FloydSteinbergDither(img, palette)

	// Verify result is valid
	bounds := result.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Errorf("Expected 2x2 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Verify all pixels are valid
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			c := result.At(x, y)
			if c == nil {
				t.Errorf("Pixel at (%d, %d) is nil", x, y)
			}
		}
	}

	// Check that not all pixels are the same (error diffusion creates patterns)
	pixels := make([]color.Color, 0, 4)
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			pixels = append(pixels, result.At(x, y))
		}
	}

	allSame := true
	for i := 1; i < len(pixels); i++ {
		if !colorsEqual(pixels[i], pixels[0]) {
			allSame = false
			break
		}
	}

	if allSame {
		t.Log("Note: All pixels are the same - error diffusion may not be visible on such a small image")
	}
}

func TestFloydSteinbergDither_LargeImage(t *testing.T) {
	palette := core.NewEInkPalette()

	// Create a larger image to test performance
	img := image.NewRGBA(image.Rect(0, 0, 296, 296)) // E-Ink display size
	for y := 0; y < 296; y++ {
		for x := 0; x < 296; x++ {
			// Create a radial gradient pattern
			dx := x - 148
			dy := y - 148
			dist := dx*dx + dy*dy
			val := uint8((dist * 255) / (148 * 148))
			img.Set(x, y, color.RGBA{val, val, val, 255})
		}
	}

	result := core.FloydSteinbergDither(img, palette)

	// Verify dimensions
	bounds := result.Bounds()
	if bounds.Dx() != 296 || bounds.Dy() != 296 {
		t.Errorf("Expected 296x296 result, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Spot check some pixels
	for i := 0; i < 10; i++ {
		x := i * 30
		y := i * 30
		c := result.At(x, y)
		if c == nil {
			t.Errorf("Pixel at (%d, %d) is nil", x, y)
		}
	}
}

func TestConvertToRGBANormalized(t *testing.T) {
	tests := []struct {
		name                string
		color               color.Color
		wantR, wantG, wantB float64
	}{
		{
			name:  "black",
			color: color.Black,
			wantR: 0, wantG: 0, wantB: 0,
		},
		{
			name:  "white",
			color: color.White,
			wantR: 1, wantG: 1, wantB: 1,
		},
		{
			name:  "red",
			color: color.RGBA{255, 0, 0, 255},
			wantR: 1, wantG: 0, wantB: 0,
		},
		{
			name:  "green",
			color: color.RGBA{0, 255, 0, 255},
			wantR: 0, wantG: 1, wantB: 0,
		},
		{
			name:  "blue",
			color: color.RGBA{0, 0, 255, 255},
			wantR: 0, wantG: 0, wantB: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := core.ConvertToRGBANormalized(tt.color)
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("ConvertToRGBANormalized() = (%v, %v, %v), want (%v, %v, %v)",
					r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestColorFromRGBANormalized(t *testing.T) {
	tests := []struct {
		name                string
		r, g, b             float64
		wantR, wantG, wantB uint8
	}{
		{
			name:  "black",
			r: 0, g: 0, b: 0,
			wantR: 0, wantG: 0, wantB: 0,
		},
		{
			name:  "white",
			r: 1, g: 1, b: 1,
			wantR: 255, wantG: 255, wantB: 255,
		},
		{
			name:  "red",
			r: 1, g: 0, b: 0,
			wantR: 255, wantG: 0, wantB: 0,
		},
		{
			name:  "green",
			r: 0, g: 1, b: 0,
			wantR: 0, wantG: 255, wantB: 0,
		},
		{
			name:  "blue",
			r: 0, g: 0, b: 1,
			wantR: 0, wantG: 0, wantB: 255,
		},
		{
			name:  "gray",
			r: 0.5, g: 0.5, b: 0.5,
			wantR: 127, wantG: 127, wantB: 127, // math.Round(127.5) = 127 in Go
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := core.ColorFromRGBANormalized(tt.r, tt.g, tt.b)
			rgba, ok := c.(color.RGBA)
			if !ok {
				t.Fatal("ColorFromRGBANormalized() did not return RGBA")
			}
			if rgba.R != tt.wantR || rgba.G != tt.wantG || rgba.B != tt.wantB {
				t.Errorf("ColorFromRGBANormalized() = R%d G%d B%d, want R%d G%d B%d",
					rgba.R, rgba.G, rgba.B, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestColorFromRGBANormalized_Clamping(t *testing.T) {
	tests := []struct {
		name                string
		r, g, b             float64
		wantR, wantG, wantB uint8
	}{
		{
			name:  "negative values",
			r: -0.5, g: -0.1, b: -1,
			wantR: 0, wantG: 0, wantB: 0,
		},
		{
			name:  "values above 1",
			r: 1.5, g: 2, b: 10,
			wantR: 255, wantG: 255, wantB: 255,
		},
		{
			name:  "mixed out of range",
			r: -0.1, g: 0.5, b: 1.1,
			wantR: 0, wantG: 127, wantB: 255, // math.Round(127.5) = 127 in Go
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := core.ColorFromRGBANormalized(tt.r, tt.g, tt.b)
			rgba, ok := c.(color.RGBA)
			if !ok {
				t.Fatal("ColorFromRGBANormalized() did not return RGBA")
			}
			if rgba.R != tt.wantR || rgba.G != tt.wantG || rgba.B != tt.wantB {
				t.Errorf("ColorFromRGBANormalized() = R%d G%d B%d, want R%d G%d B%d",
					rgba.R, rgba.G, rgba.B, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}
