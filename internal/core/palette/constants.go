// Package palette provides the 6-color E-Ink display palette and color utilities.
package palette

import (
	"image/color"
)

// PaletteIndex represents the index of a color in the E-Ink palette.
type PaletteIndex byte

const (
	// IndexBlack is the palette index for black (0, 0, 0)
	IndexBlack PaletteIndex = 0
	// IndexWhite is the palette index for white (255, 255, 255)
	IndexWhite PaletteIndex = 1
	// IndexGreen is the palette index for green (0, 255, 0)
	IndexGreen PaletteIndex = 2
	// IndexBlue is the palette index for blue (0, 0, 255)
	IndexBlue PaletteIndex = 3
	// IndexRed is the palette index for red (255, 0, 0)
	IndexRed PaletteIndex = 4
	// IndexYellow is the palette index for yellow (255, 255, 0)
	IndexYellow PaletteIndex = 5
)

// EInkColors returns the 6 supported E-Ink colors in palette index order.
func EInkColors() []color.Color {
	return []color.Color{
		ColorBlack,
		ColorWhite,
		ColorGreen,
		ColorBlue,
		ColorRed,
		ColorYellow,
	}
}

// PaletteSize returns the number of colors in the E-Ink palette.
func PaletteSize() int {
	return 6
}

// Color constants for the E-Ink palette.
var (
	// ColorBlack is black color (0, 0, 0)
	ColorBlack = color.Black
	// ColorWhite is white color (255, 255, 255)
	ColorWhite = color.White
	// ColorGreen is green color (0, 255, 0)
	ColorGreen = color.RGBA{0, 255, 0, 255}
	// ColorBlue is blue color (0, 0, 255)
	ColorBlue = color.RGBA{0, 0, 255, 255}
	// ColorRed is red color (255, 0, 0)
	ColorRed = color.RGBA{255, 0, 0, 255}
	// ColorYellow is yellow color (255, 255, 0)
	ColorYellow = color.RGBA{255, 255, 0, 255}
)

// IndexForColor returns the palette index for a given palette color.
// Returns -1 if the color is not in the palette.
func IndexForColor(c color.Color) int {
	switch c {
	case ColorBlack:
		return int(IndexBlack)
	case ColorWhite:
		return int(IndexWhite)
	case ColorGreen:
		return int(IndexGreen)
	case ColorBlue:
		return int(IndexBlue)
	case ColorRed:
		return int(IndexRed)
	case ColorYellow:
		return int(IndexYellow)
	default:
		return -1
	}
}

// ColorForIndex returns the color for the given palette index.
// Returns nil if the index is out of range.
func ColorForIndex(idx PaletteIndex) color.Color {
	switch idx {
	case IndexBlack:
		return ColorBlack
	case IndexWhite:
		return ColorWhite
	case IndexGreen:
		return ColorGreen
	case IndexBlue:
		return ColorBlue
	case IndexRed:
		return ColorRed
	case IndexYellow:
		return ColorYellow
	default:
		return nil
	}
}
