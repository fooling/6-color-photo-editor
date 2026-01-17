// Package pipeline provides image filtering and processing functionality for E-Ink displays.
//
// The package implements a flexible filter pipeline that can be used to:
//   - Resize images to E-Ink display dimensions
//   - Adjust brightness, contrast, and saturation
//   - Apply Floyd-Steinberg dithering for 6-color output
//   - Capture intermediate processing steps for visualization
package pipeline

import (
	"image"
)

// Filter represents a transform that can be applied to an image.
// Filters are the basic building blocks of the image processing pipeline.
type Filter interface {
	// Name returns the filter's name.
	Name() string

	// Apply applies the filter to an image and returns the result.
	Apply(img image.Image) (image.Image, error)
}

// ProcessConfig holds configuration for image processing.
// All fields are optional; zero values mean no change.
type ProcessConfig struct {
	// Width and Height specify the target dimensions.
	// If one dimension is zero, the aspect ratio is maintained.
	// If both are zero, no resizing is performed.
	Width      int
	Height     int
	// Crop rectangle (relative coordinates 0.0-1.0). If all zeros, no cropping.
	CropX      float64 // Top-left X (0.0 to 1.0)
	CropY      float64 // Top-left Y (0.0 to 1.0)
	CropWidth  float64 // Crop width (0.0 to 1.0)
	CropHeight float64 // Crop height (0.0 to 1.0)
	// Output format: "png" or "bmp"
	OutputFormat string
	Brightness   float64 // Brightness adjustment from -1.0 to 1.0
	Contrast     float64 // Contrast adjustment (1.0 = no change, >1.0 = higher contrast)
	Saturation   float64 // Saturation adjustment (1.0 = no change, 0.0 = grayscale, >1.0 = oversaturated)
	Dither       bool    // Enable Floyd-Steinberg dithering
	EnhancerName string  // Name of the enhancer to use (empty = basic with brightness/contrast/saturation)
}

// DefaultConfig returns a default processing configuration.
// The default enables dithering with no other adjustments.
func DefaultConfig() *ProcessConfig {
	return &ProcessConfig{
		OutputFormat: "png",
		Brightness:   0.0,
		Contrast:     1.0,
		Saturation:   1.0,
		Dither:       true,
	}
}

// ProcessingStep represents a step in the pipeline for visualization.
// It contains the name of the filter and the resulting image after applying that filter.
type ProcessingStep struct {
	Name  string
	Image image.Image
}

// ProcessingResult contains the final result and optional intermediate steps.
// If step capture was enabled during pipeline creation, Steps will contain
// the intermediate images from key processing stages.
type ProcessingResult struct {
	Final  image.Image           // The final processed image
	Steps  []ProcessingStep      // Intermediate processing steps (if captured)
	Config *ProcessConfig        // The configuration used for processing
}
