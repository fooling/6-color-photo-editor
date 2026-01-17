package pipeline

import (
	"github.com/fooling/6-color-editor/internal/core"
	"image"
)

// DitherFilter applies Floyd-Steinberg dithering using the 6-color E-Ink palette
type DitherFilter struct {
	palette *core.EInkPalette
}

// NewDitherFilter creates a new dither filter
func NewDitherFilter() *DitherFilter {
	return &DitherFilter{
		palette: core.NewEInkPalette(),
	}
}

// Name returns the filter name
func (f *DitherFilter) Name() string {
	return "Dither"
}

// Apply applies dithering to the image
func (f *DitherFilter) Apply(img image.Image) (image.Image, error) {
	result := core.FloydSteinbergDither(img, f.palette)
	return result, nil
}
