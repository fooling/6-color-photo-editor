package pipeline

import (
	"image"
)

// CropFilter crops an image to the specified rectangle.
// The crop rectangle is specified as relative coordinates (0.0 to 1.0)
// to support responsive frontend cropping UI.
type CropFilter struct {
	// X, Y are the top-left corner of the crop rectangle (0.0 to 1.0)
	X, Y float64
	// Width, Height are the dimensions of the crop rectangle (0.0 to 1.0)
	Width, Height float64
}

// NewCropFilter creates a new crop filter.
// All parameters are relative (0.0 to 1.0) for responsive UI.
//
// Example: Crop the center 50% of an image:
//   f := NewCropFilter(0.25, 0.25, 0.5, 0.5)
func NewCropFilter(x, y, width, height float64) *CropFilter {
	return &CropFilter{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// Name returns the filter name
func (f *CropFilter) Name() string {
	return "Crop"
}

// Apply crops the image to the specified rectangle
func (f *CropFilter) Apply(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Convert relative coordinates to absolute pixels
	x := int(float64(origWidth) * f.X)
	y := int(float64(origHeight) * f.Y)
	width := int(float64(origWidth) * f.Width)
	height := int(float64(origHeight) * f.Height)

	// Clamp values to image bounds
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+width > origWidth {
		width = origWidth - x
	}
	if y+height > origHeight {
		height = origHeight - y
	}

	// Create cropped image with bounds starting at (0, 0)
	cropRect := image.Rect(0, 0, width, height)
	cropped := image.NewRGBA(cropRect)

	// Copy pixels from source to cropped image
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			cropped.Set(dx, dy, img.At(x+dx, y+dy))
		}
	}

	return cropped, nil
}

// CropToAspectFilter crops an image to a target aspect ratio.
// It will crop the minimum area needed to achieve the target ratio.
type CropToAspectFilter struct {
	TargetWidth  int
	TargetHeight int
}

// NewCropToAspectFilter creates a new aspect ratio crop filter.
func NewCropToAspectFilter(targetWidth, targetHeight int) *CropToAspectFilter {
	return &CropToAspectFilter{
		TargetWidth:  targetWidth,
		TargetHeight: targetHeight,
	}
}

// Name returns the filter name
func (f *CropToAspectFilter) Name() string {
	return "CropAspect"
}

// Apply crops the image to the target aspect ratio
func (f *CropToAspectFilter) Apply(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate aspect ratios
	targetAspect := float64(f.TargetWidth) / float64(f.TargetHeight)
	currentAspect := float64(origWidth) / float64(origHeight)

	var srcX, srcY, srcWidth, srcHeight int

	if currentAspect > targetAspect {
		// Image is wider than target - crop sides
		srcWidth = int(float64(origHeight) * targetAspect)
		srcHeight = origHeight
		srcX = (origWidth - srcWidth) / 2
		srcY = 0
	} else {
		// Image is taller than target - crop top/bottom
		srcWidth = origWidth
		srcHeight = int(float64(origWidth) / targetAspect)
		srcX = 0
		srcY = (origHeight - srcHeight) / 2
	}

	// Create cropped image with bounds starting at (0, 0)
	cropRect := image.Rect(0, 0, srcWidth, srcHeight)
	cropped := image.NewRGBA(cropRect)

	// Copy pixels from source to cropped image
	for dy := 0; dy < srcHeight; dy++ {
		for dx := 0; dx < srcWidth; dx++ {
			cropped.Set(dx, dy, img.At(srcX+dx, srcY+dy))
		}
	}

	return cropped, nil
}
