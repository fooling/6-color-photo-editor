package pipeline

import (
	"image"
	"image/color"
	"math"
)

// EnhanceFilter adjusts brightness, contrast, and saturation
type EnhanceFilter struct {
	brightness float64 // -1.0 to 1.0
	contrast   float64 // 0.0 to 2.0+
	saturation float64 // 0.0 to 2.0+
}

// NewEnhanceFilter creates a new enhancement filter
func NewEnhanceFilter(brightness, contrast, saturation float64) *EnhanceFilter {
	return &EnhanceFilter{
		brightness: brightness,
		contrast:   contrast,
		saturation: saturation,
	}
}

// Name returns the filter name
func (f *EnhanceFilter) Name() string {
	return "Enhance"
}

// Apply applies the enhancements to the image
func (f *EnhanceFilter) Apply(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			result.Set(x, y, f.applyEnhancement(c))
		}
	}

	return result, nil
}

// applyEnhancement applies all enhancements to a single color
func (f *EnhanceFilter) applyEnhancement(c color.Color) color.Color {
	r, g, b, a := c.RGBA()

	// Convert to 0-1 range
	rNorm := float64(r>>8) / 255.0
	gNorm := float64(g>>8) / 255.0
	bNorm := float64(b>>8) / 255.0

	// Apply brightness
	if f.brightness != 0 {
		rNorm = clamp(rNorm + f.brightness)
		gNorm = clamp(gNorm + f.brightness)
		bNorm = clamp(bNorm + f.brightness)
	}

	// Apply contrast
	if f.contrast != 1.0 {
		rNorm = clamp((rNorm-0.5)*f.contrast + 0.5)
		gNorm = clamp((gNorm-0.5)*f.contrast + 0.5)
		bNorm = clamp((bNorm-0.5)*f.contrast + 0.5)
	}

	// Apply saturation (using HSL-style approach)
	if f.saturation != 1.0 {
		// Calculate luminance (perceptual brightness)
		lum := 0.299*rNorm + 0.587*gNorm + 0.114*bNorm

		// Interpolate between luminance and original color based on saturation
		rNorm = clamp(lum + (rNorm-lum)*f.saturation)
		gNorm = clamp(lum + (gNorm-lum)*f.saturation)
		bNorm = clamp(lum + (bNorm-lum)*f.saturation)
	}

	return color.RGBA{
		R: uint8(math.Round(rNorm * 255)),
		G: uint8(math.Round(gNorm * 255)),
		B: uint8(math.Round(bNorm * 255)),
		A: uint8(a >> 8),
	}
}

// clamp ensures a value is within [0, 1]
func clamp(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
