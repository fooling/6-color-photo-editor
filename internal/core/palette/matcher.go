package palette

import (
	"image/color"
)

// Matcher finds the closest E-Ink palette color for any given color.
type Matcher struct {
	colors []color.Color
}

// NewMatcher creates a new color matcher initialized with the E-Ink palette.
func NewMatcher() *Matcher {
	return &Matcher{
		colors: EInkColors(),
	}
}

// FindClosestColor finds the nearest palette color using Euclidean distance in RGB space.
func (m *Matcher) FindClosestColor(c color.Color) color.Color {
	r, g, b, _ := c.RGBA()

	// Convert from 16-bit to 8-bit
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)

	minDist := uint32(0xFFFFFFFF)
	var closest color.Color

	for _, pc := range m.colors {
		pr, pg, pb, _ := pc.RGBA()
		pr8 := uint8(pr >> 8)
		pg8 := uint8(pg >> 8)
		pb8 := uint8(pb >> 8)

		// Euclidean distance squared (no sqrt needed for comparison)
		dr := uint32(r8) - uint32(pr8)
		dg := uint32(g8) - uint32(pg8)
		db := uint32(b8) - uint32(pb8)
		dist := dr*dr + dg*dg + db*db

		if dist < minDist {
			minDist = dist
			closest = pc
		}
	}

	return closest
}

// FindClosestIndex finds the nearest palette color index using Euclidean distance in RGB space.
func (m *Matcher) FindClosestIndex(c color.Color) PaletteIndex {
	closest := m.FindClosestColor(c)
	return PaletteIndex(IndexForColor(closest))
}
