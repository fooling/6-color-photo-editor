package pipeline

import (
	"image"
)

// Pipeline manages a sequence of filters to apply to images.
// Filters are executed in the order they were added.
//
// The pipeline supports optional step capture, which stores intermediate
// results after key filters (Resize, Enhance, Dither) for visualization.
type Pipeline struct {
	filters         []Filter
	captureSteps    bool
	captureAllSteps bool
}

// NewPipeline creates a new pipeline with optional step capture.
//
// The captureSteps parameter controls whether intermediate processing
// results are stored. When enabled, steps are captured after filters
// named "Resize", "Enhance", and "Dither".
//
// Example:
//   p := pipeline.NewPipeline(true)
//   p.AddFilter(pipeline.NewResizeFilter(800, 480))
//   p.AddFilter(pipeline.NewDitherFilter())
//   result := p.Process(img)
func NewPipeline(captureSteps bool) *Pipeline {
	return &Pipeline{
		filters:      make([]Filter, 0),
		captureSteps: captureSteps,
	}
}

// AddFilter appends a filter to the pipeline.
// Returns the pipeline for method chaining.
//
// Example:
//   p.AddFilter(NewResizeFilter(100, 100))
//    .AddFilter(NewEnhanceFilter(0.1, 1.0, 1.0))
func (p *Pipeline) AddFilter(f Filter) *Pipeline {
	p.filters = append(p.filters, f)
	return p
}

// Clear removes all filters from the pipeline.
// Returns the pipeline for method chaining.
func (p *Pipeline) Clear() *Pipeline {
	p.filters = make([]Filter, 0)
	return p
}

// Process executes all filters in sequence on the input image.
// If an error occurs during filter execution, processing stops and
// the current result is returned along with any captured steps.
//
// Returns a ProcessingResult containing the final image and any
// intermediate steps that were captured.
//
// Example:
//   result := p.Process(sourceImg)
//   finalImg := result.Final
func (p *Pipeline) Process(img image.Image) *ProcessingResult {
	result := &ProcessingResult{
		Config: &ProcessConfig{},
	}

	current := img
	steps := make([]ProcessingStep, 0)

	// Capture original if capturing all steps
	if p.captureSteps && p.captureAllSteps {
		steps = append(steps, ProcessingStep{
			Name:  "Original",
			Image: current,
		})
	}

	for _, f := range p.filters {
		var err error
		current, err = f.Apply(current)
		if err != nil {
			// Return what we have so far on error
			result.Final = current
			result.Steps = steps
			return result
		}

		// Capture step after filter
		if p.captureSteps {
			// Always capture dither step
			shouldCapture := p.captureAllSteps || f.Name() == "Dither"

			// Also capture after key transformation steps
			if !shouldCapture {
				switch f.Name() {
				case "Resize", "Enhance":
					shouldCapture = true
				}
			}

			if shouldCapture {
				steps = append(steps, ProcessingStep{
					Name:  f.Name(),
					Image: current,
				})
			}
		}
	}

	result.Final = current
	result.Steps = steps
	return result
}

// ProcessWithConfig processes an image using the specified configuration.
// This is a convenience method that builds the pipeline from config.
//
// Filters are added in the correct order:
//   1. Crop (if crop rectangle specified)
//   2. Resize (if dimensions specified)
//   3. Enhance (using selected enhancer or basic with brightness/contrast/saturation)
//   4. Dither (if enabled)
//
// Example:
//   config := &pipeline.ProcessConfig{
//       Width: 800,
//       Height: 480,
//       OutputFormat: "bmp",
//       Dither: true,
//   }
//   result := pipeline.ProcessWithConfig(img, config, true)
func ProcessWithConfig(img image.Image, config *ProcessConfig, captureSteps bool) *ProcessingResult {
	p := NewPipeline(captureSteps)

	// Add filters in the correct order

	// 1. Crop first (if specified)
	if config.CropWidth > 0 && config.CropHeight > 0 {
		// Use user-specified crop rectangle
		p.AddFilter(NewCropFilter(config.CropX, config.CropY, config.CropWidth, config.CropHeight))
	} else if config.Width > 0 && config.Height > 0 {
		// No crop specified, but target dimensions provided.
		// Crop to target aspect ratio first (center crop), matching frontend behavior.
		p.AddFilter(NewCropToAspectFilter(config.Width, config.Height))
	}

	// 2. Resize to target dimensions
	if config.Width > 0 || config.Height > 0 {
		p.AddFilter(NewResizeFilter(config.Width, config.Height))
	}

	// 3. Enhance - use selected enhancer or basic with sliders
	if config.EnhancerName != "" && config.EnhancerName != "basic" {
		// Use the selected enhancer from registry
		if enhancer, ok := GetEnhancer(config.EnhancerName); ok {
			p.AddFilter(&enhancerFilterAdapter{enhancer: enhancer})
		} else {
			// Fallback to basic if enhancer not found
			p.AddFilter(NewEnhanceFilter(config.Brightness, config.Contrast, config.Saturation))
		}
	} else {
		// Use basic enhancer with slider values
		p.AddFilter(NewEnhanceFilter(config.Brightness, config.Contrast, config.Saturation))
	}

	// 4. Dither to 6-color palette
	if config.Dither {
		p.AddFilter(NewDitherFilter())
	}

	result := p.Process(img)
	result.Config = config
	return result
}

// enhancerFilterAdapter wraps an Enhancer to implement the Filter interface
type enhancerFilterAdapter struct {
	enhancer Enhancer
}

func (a *enhancerFilterAdapter) Name() string {
	return "Enhance"
}

func (a *enhancerFilterAdapter) Apply(img image.Image) (image.Image, error) {
	return a.enhancer.Apply(img)
}
