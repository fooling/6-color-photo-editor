package pipeline

import (
	"github.com/disintegration/imaging"
	"image"
)

// ResizeFilter resizes an image to the specified dimensions
// If width or height is 0, it maintains aspect ratio
type ResizeFilter struct {
	width  int
	height int
}

// NewResizeFilter creates a new resize filter
func NewResizeFilter(width, height int) *ResizeFilter {
	return &ResizeFilter{
		width:  width,
		height: height,
	}
}

// Name returns the filter name
func (f *ResizeFilter) Name() string {
	return "Resize"
}

// Apply resizes the image
func (f *ResizeFilter) Apply(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// If both dimensions are specified, resize to exact size
	if f.width > 0 && f.height > 0 {
		return imaging.Resize(img, f.width, f.height, imaging.Lanczos), nil
	}

	// If only width is specified, calculate height maintaining aspect ratio
	if f.width > 0 {
		newHeight := (origHeight * f.width) / origWidth
		return imaging.Resize(img, f.width, newHeight, imaging.Lanczos), nil
	}

	// If only height is specified, calculate width maintaining aspect ratio
	if f.height > 0 {
		newWidth := (origWidth * f.height) / origHeight
		return imaging.Resize(img, newWidth, f.height, imaging.Lanczos), nil
	}

	// No resize needed
	return img, nil
}
