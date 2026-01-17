// Package core provides core image processing functionality for E-Ink displays.
//
// The package includes functionality for:
//   - 6-color E-Ink palette management
//   - Floyd-Steinberg dithering for color reduction
//   - Color conversion utilities
package core

import (
	"image/color"

	"github.com/fooling/6-color-editor/internal/core/palette"
)

// EInkPalette represents the 6-color E-Ink display palette.
// The palette consists of: Black, White, Green, Blue, Red, and Yellow.
type EInkPalette struct {
	matcher *palette.Matcher
}

// NewEInkPalette creates a new E-Ink palette initialized with the standard 6-color set.
//
// Example:
//   p := core.NewEInkPalette()
//   closest := p.FindClosestColor(color.RGBA{128, 128, 128, 255})
func NewEInkPalette() *EInkPalette {
	return &EInkPalette{
		matcher: palette.NewMatcher(),
	}
}

// Colors returns the 6 supported E-Ink colors.
// The colors are returned in the order: Black, White, Green, Blue, Red, Yellow.
func (p *EInkPalette) Colors() []color.Color {
	return palette.EInkColors()
}

// FindClosestColor finds the nearest palette color using Euclidean distance in RGB space.
// This is useful for determining which E-Ink color best represents a given input color.
//
// The algorithm calculates the squared Euclidean distance:
//   distance² = (r1-r2)² + (g1-g2)² + (b1-b2)²
//
// Example:
//   p := core.NewEInkPalette()
//   closest := p.FindClosestColor(color.RGBA{200, 100, 50, 255})
func (p *EInkPalette) FindClosestColor(c color.Color) color.Color {
	return p.matcher.FindClosestColor(c)
}

// ConvertToRGBANormalized converts a color to normalized 0-1 float64 RGB values.
// The alpha channel is discarded.
//
// The conversion formula is:
//   normalized = 16-bit value / 65535.0
//
// Example:
//   r, g, b := core.ConvertToRGBANormalized(color.RGBA{128, 128, 128, 255})
//   // r, g, b ≈ 0.5
func ConvertToRGBANormalized(c color.Color) (r, g, b float64) {
	rUint, gUint, bUint, _ := c.RGBA()
	r = float64(rUint) / 65535.0
	g = float64(gUint) / 65535.0
	b = float64(bUint) / 65535.0
	return
}

// ColorFromRGBANormalized creates a color from normalized 0-1 float64 RGB values.
// Values outside the [0, 1] range are clamped. The resulting color is fully opaque (alpha = 255).
//
// Example:
//   c := core.ColorFromRGBANormalized(0.5, 0.5, 0.5) // Returns gray
func ColorFromRGBANormalized(r, g, b float64) color.Color {
	// Clamp values to [0, 1]
	clamp := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}

	r = clamp(r)
	g = clamp(g)
	b = clamp(b)

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}
