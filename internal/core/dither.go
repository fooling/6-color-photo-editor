package core

import (
	"image"
	"image/color"
	"math"
)

// FloydSteinbergDither applies Floyd-Steinberg dithering to an image
// using the 6-color E-Ink palette.
//
// Floyd-Steinberg dithering is an error diffusion algorithm that reduces
// color banding by distributing quantization errors to neighboring pixels.
// This produces higher quality results than simple nearest-color matching.
//
// The error is distributed using the following weights:
//   - Right pixel (current row, next column): 7/16
//   - Bottom-left pixel (next row, previous column): 3/16
//   - Bottom pixel (next row, same column): 5/16
//   - Bottom-right pixel (next row, next column): 1/16
//
// Parameters:
//   - img: The source image to dither
//   - palette: The E-Ink palette to use for color quantization
//
// Returns:
//   - A new image with dithering applied
//
// Example:
//   p := core.NewEInkPalette()
//   dithered := core.FloydSteinbergDither(srcImg, p)
func FloydSteinbergDither(img image.Image, palette *EInkPalette) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new image for the dithered result
	result := image.NewRGBA(bounds)

	// Working buffer with float64 precision for error diffusion
	// Using [3][]float64 layout: [height][width*3] for RGB
	buffer := make([][]float64, height)
	for y := 0; y < height; y++ {
		buffer[y] = make([]float64, width*3)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			buffer[y][x*3] = float64(r>>8)   // R
			buffer[y][x*3+1] = float64(g>>8) // G
			buffer[y][x*3+2] = float64(b>>8) // B
		}
	}

	// Process each pixel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := x * 3

			// Get current pixel values
			oldR := buffer[y][offset]
			oldG := buffer[y][offset+1]
			oldB := buffer[y][offset+2]

			// Clamp to valid range
			oldR = clamp(oldR)
			oldG = clamp(oldG)
			oldB = clamp(oldB)

			// Find closest palette color
			oldColor := color.RGBA{
				R: uint8(oldR),
				G: uint8(oldG),
				B: uint8(oldB),
				A: 255,
			}
			newColor := palette.FindClosestColor(oldColor)

			// Set the pixel in result
			result.Set(x, y, newColor)

			// Calculate quantization error
			newR, newG, newB, _ := newColor.RGBA()
			errR := oldR - float64(newR>>8)
			errG := oldG - float64(newG>>8)
			errB := oldB - float64(newB>>8)

			// Distribute error to neighboring pixels
			// Right pixel (7/16)
			if x+1 < width {
				distributeError(buffer, y, x+1, errR, errG, errB, 7.0/16.0)
			}

			// Bottom-left pixel (3/16)
			if y+1 < height && x-1 >= 0 {
				distributeError(buffer, y+1, x-1, errR, errG, errB, 3.0/16.0)
			}

			// Bottom pixel (5/16)
			if y+1 < height {
				distributeError(buffer, y+1, x, errR, errG, errB, 5.0/16.0)
			}

			// Bottom-right pixel (1/16)
			if y+1 < height && x+1 < width {
				distributeError(buffer, y+1, x+1, errR, errG, errB, 1.0/16.0)
			}
		}
	}

	return result
}

// distributeError adds a portion of the quantization error to a pixel.
func distributeError(buffer [][]float64, y, x int, errR, errG, errB, factor float64) {
	offset := x * 3
	buffer[y][offset] += errR * factor
	buffer[y][offset+1] += errG * factor
	buffer[y][offset+2] += errB * factor
}

// clamp ensures a value is within [0, 255].
func clamp(v float64) float64 {
	return math.Max(0, math.Min(255, v))
}
