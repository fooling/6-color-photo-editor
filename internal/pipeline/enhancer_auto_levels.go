package pipeline

import (
	"image"
	"image/color"
	"math"
)

// AutoLevelsEnhancer automatically adjusts image levels based on histogram analysis.
// It stretches the histogram to use the full dynamic range, improving contrast
// in images that don't use the full brightness range.
type AutoLevelsEnhancer struct {
	clipPercent float64 // Percentage of pixels to clip at each end (0.0 to 5.0)
}

// NewAutoLevelsEnhancer creates a new auto levels enhancer
// clipPercent specifies how many percent of pixels to ignore at each end
// of the histogram (helps ignore outliers). Recommended: 0.5 to 2.0
func NewAutoLevelsEnhancer(clipPercent float64) *AutoLevelsEnhancer {
	return &AutoLevelsEnhancer{
		clipPercent: clipPercent,
	}
}

// Name returns the enhancer identifier
func (e *AutoLevelsEnhancer) Name() string {
	return "auto_levels"
}

// DisplayName returns the human-readable name
func (e *AutoLevelsEnhancer) DisplayName() string {
	return "Auto Levels"
}

// Description returns a brief description
func (e *AutoLevelsEnhancer) Description() string {
	return "Automatic histogram stretching for improved contrast"
}

// Apply applies auto levels adjustment to the image
func (e *AutoLevelsEnhancer) Apply(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	totalPixels := width * height

	// Build histograms for each channel
	histR := make([]int, 256)
	histG := make([]int, 256)
	histB := make([]int, 256)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			histR[r>>8]++
			histG[g>>8]++
			histB[b>>8]++
		}
	}

	// Calculate clip threshold
	clipCount := int(float64(totalPixels) * e.clipPercent / 100.0)

	// Find min/max for each channel after clipping
	minR, maxR := findClippedRange(histR, clipCount)
	minG, maxG := findClippedRange(histG, clipCount)
	minB, maxB := findClippedRange(histB, clipCount)

	// Create lookup tables
	lutR := createStretchLUT(minR, maxR)
	lutG := createStretchLUT(minG, maxG)
	lutB := createStretchLUT(minB, maxB)

	// Apply transformation
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			result.Set(x, y, color.RGBA{
				R: lutR[r>>8],
				G: lutG[g>>8],
				B: lutB[b>>8],
				A: uint8(a >> 8),
			})
		}
	}

	return result, nil
}

// findClippedRange finds the min and max values after ignoring clipCount pixels
func findClippedRange(hist []int, clipCount int) (min, max int) {
	// Find minimum (from left, skipping clipCount pixels)
	count := 0
	min = 0
	for i := 0; i < 256; i++ {
		count += hist[i]
		if count > clipCount {
			min = i
			break
		}
	}

	// Find maximum (from right, skipping clipCount pixels)
	count = 0
	max = 255
	for i := 255; i >= 0; i-- {
		count += hist[i]
		if count > clipCount {
			max = i
			break
		}
	}

	// Ensure min < max
	if min >= max {
		min = 0
		max = 255
	}

	return min, max
}

// createStretchLUT creates a lookup table for stretching values from [min,max] to [0,255]
func createStretchLUT(min, max int) []uint8 {
	lut := make([]uint8, 256)
	rangeVal := float64(max - min)
	if rangeVal < 1 {
		rangeVal = 1
	}

	for i := 0; i < 256; i++ {
		if i <= min {
			lut[i] = 0
		} else if i >= max {
			lut[i] = 255
		} else {
			lut[i] = uint8(math.Round(float64(i-min) / rangeVal * 255))
		}
	}

	return lut
}

func init() {
	// Register with default 1% clip
	RegisterEnhancer(NewAutoLevelsEnhancer(1.0))
}
