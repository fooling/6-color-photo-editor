// Package encoder provides encoding functionality for E-Ink display formats.
//
// The package implements a binary encoding format for 6-color E-Ink displays.
// The encoded format consists of:
//   - 2 bytes: width (big endian uint16)
//   - 2 bytes: height (big endian uint16)
//   - N bytes: pixel data (1 byte per pixel, palette index 0-5)
//
// Palette indices:
//   0: Black
//   1: White
//   2: Green
//   3: Blue
//   4: Red
//   5: Yellow
package encoder

import (
	"encoding/binary"
	"image"
	"image/color"
	"log"

	"github.com/fooling/6-color-editor/internal/core/palette"
)

// EInk encodes images to the binary format expected by E-Ink displays.
// Colors not in the palette are automatically matched to the nearest palette color.
type EInk struct {
	matcher *palette.Matcher
}

// NewEInk creates a new E-Ink encoder.
func NewEInk() *EInk {
	return &EInk{
		matcher: palette.NewMatcher(),
	}
}

// Encode converts an image to the binary format expected by the E-Ink display.
// The format is:
//   - 2 bytes: width (big endian)
//   - 2 bytes: height (big endian)
//   - N bytes: pixel data (1 byte per pixel, palette index 0-5)
//
// If a pixel's color is not exactly in the palette, the closest palette
// color is found using Euclidean distance in RGB space.
//
// Example:
//   e := encoder.NewEInk()
//   data, err := e.Encode(img)
//   if err != nil {
//       log.Fatal(err)
//   }
func (e *EInk) Encode(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Buffer to hold the pixel data
	// Using 1 byte per pixel (we only need 3 bits for 6 colors)
	data := make([]byte, 4+width*height) // 4 bytes header + pixel data

	// Write header: width (2 bytes) + height (2 bytes)
	binary.BigEndian.PutUint16(data[0:2], uint16(width))
	binary.BigEndian.PutUint16(data[2:4], uint16(height))

	// Convert each pixel to its palette index
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.At(x, y)
			idx := e.findIndex(c)
			data[4+y*width+x] = byte(idx)
		}
	}

	log.Printf("[Encoder] Encoded %dx%d image, total %d bytes", width, height, len(data))
	log.Printf("[Encoder] Header: width=%d, height=%d", binary.BigEndian.Uint16(data[0:2]), binary.BigEndian.Uint16(data[2:4]))
	// Log first 16 bytes of pixel data for debugging
	pixelDataLen := len(data) - 4
	previewLen := 16
	if pixelDataLen < previewLen {
		previewLen = pixelDataLen
	}
	log.Printf("[Encoder] First %d bytes of pixel data: %v", previewLen, data[4:4+previewLen])

	return data, nil
}

// findIndex returns the palette index for a given color.
// If the color is not exactly in the palette, it finds the closest match.
func (e *EInk) findIndex(c color.Color) palette.PaletteIndex {
	// First try to find exact match
	if idx := palette.IndexForColor(c); idx >= 0 {
		return palette.PaletteIndex(idx)
	}
	// Otherwise find closest match
	return e.matcher.FindClosestIndex(c)
}
