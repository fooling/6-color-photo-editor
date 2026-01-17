package pipeline

import (
	"image"
	"image/color"
	"math"
)

// BasicEnhancer provides standard brightness, contrast, and saturation adjustments.
// This is the default enhancer that applies simple linear transformations.
type BasicEnhancer struct {
	brightness float64 // -1.0 to 1.0
	contrast   float64 // 0.0 to 2.0+
	saturation float64 // 0.0 to 2.0+
}

// NewBasicEnhancer creates a new basic enhancer with the specified parameters
func NewBasicEnhancer(brightness, contrast, saturation float64) *BasicEnhancer {
	return &BasicEnhancer{
		brightness: brightness,
		contrast:   contrast,
		saturation: saturation,
	}
}

// Name returns the enhancer identifier
func (e *BasicEnhancer) Name() string {
	return "basic"
}

// DisplayName returns the human-readable name
func (e *BasicEnhancer) DisplayName() string {
	return "Basic"
}

// Description returns a brief description
func (e *BasicEnhancer) Description() string {
	return "Standard brightness, contrast, and saturation adjustments"
}

// Apply applies the enhancement to the image
func (e *BasicEnhancer) Apply(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			result.Set(x, y, e.applyEnhancement(c))
		}
	}

	return result, nil
}

// applyEnhancement applies all enhancements to a single color
func (e *BasicEnhancer) applyEnhancement(c color.Color) color.Color {
	r, g, b, a := c.RGBA()

	// Convert to 0-1 range
	rNorm := float64(r>>8) / 255.0
	gNorm := float64(g>>8) / 255.0
	bNorm := float64(b>>8) / 255.0

	// Apply brightness
	if e.brightness != 0 {
		rNorm = clampFloat(rNorm + e.brightness)
		gNorm = clampFloat(gNorm + e.brightness)
		bNorm = clampFloat(bNorm + e.brightness)
	}

	// Apply contrast
	if e.contrast != 1.0 {
		rNorm = clampFloat((rNorm-0.5)*e.contrast + 0.5)
		gNorm = clampFloat((gNorm-0.5)*e.contrast + 0.5)
		bNorm = clampFloat((bNorm-0.5)*e.contrast + 0.5)
	}

	// Apply saturation (using HSL-style approach)
	if e.saturation != 1.0 {
		// Calculate luminance (perceptual brightness)
		lum := 0.299*rNorm + 0.587*gNorm + 0.114*bNorm

		// Interpolate between luminance and original color based on saturation
		rNorm = clampFloat(lum + (rNorm-lum)*e.saturation)
		gNorm = clampFloat(lum + (gNorm-lum)*e.saturation)
		bNorm = clampFloat(lum + (bNorm-lum)*e.saturation)
	}

	return color.RGBA{
		R: uint8(math.Round(rNorm * 255)),
		G: uint8(math.Round(gNorm * 255)),
		B: uint8(math.Round(bNorm * 255)),
		A: uint8(a >> 8),
	}
}

// clampFloat ensures a value is within [0, 1]
func clampFloat(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

func init() {
	// Register a default basic enhancer with neutral settings
	RegisterEnhancer(NewBasicEnhancer(0, 1, 1))
}
