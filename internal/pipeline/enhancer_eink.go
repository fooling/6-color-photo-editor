package pipeline

import (
	"image"
	"image/color"
	"math"
)

// EInk6ColorPalette defines the 6 colors available on the E-Ink display
var EInk6ColorPalette = []color.RGBA{
	{R: 0, G: 0, B: 0, A: 255},       // Black
	{R: 255, G: 255, B: 255, A: 255}, // White
	{R: 255, G: 0, B: 0, A: 255},     // Red
	{R: 0, G: 255, B: 0, A: 255},     // Green
	{R: 0, G: 0, B: 255, A: 255},     // Blue
	{R: 255, G: 255, B: 0, A: 255},   // Yellow
}

// EInkOptimizedEnhancer optimizes images specifically for 6-color E-Ink displays.
// It enhances color separation, applies slight sharpening, and pushes colors
// toward the available palette for better dithering results.
type EInkOptimizedEnhancer struct {
	colorBoost   float64 // How much to boost color separation (0.0 to 1.0)
	sharpness    float64 // Sharpening strength (0.0 to 1.0)
	contrastGain float64 // Additional contrast (1.0 = no change)
}

// NewEInkOptimizedEnhancer creates a new E-Ink optimized enhancer
func NewEInkOptimizedEnhancer(colorBoost, sharpness, contrastGain float64) *EInkOptimizedEnhancer {
	return &EInkOptimizedEnhancer{
		colorBoost:   colorBoost,
		sharpness:    sharpness,
		contrastGain: contrastGain,
	}
}

// Name returns the enhancer identifier
func (e *EInkOptimizedEnhancer) Name() string {
	return "eink_optimized"
}

// DisplayName returns the human-readable name
func (e *EInkOptimizedEnhancer) DisplayName() string {
	return "E-Ink Optimized"
}

// Description returns a brief description
func (e *EInkOptimizedEnhancer) Description() string {
	return "Optimized for 6-color E-Ink displays with enhanced color separation"
}

// Apply applies E-Ink optimization to the image
func (e *EInkOptimizedEnhancer) Apply(img image.Image) (image.Image, error) {
	bounds := img.Bounds()

	// First pass: enhance colors and contrast
	enhanced := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			enhanced.Set(x, y, e.enhancePixel(c))
		}
	}

	// Second pass: apply sharpening if enabled
	if e.sharpness > 0 {
		return e.applySharpen(enhanced), nil
	}

	return enhanced, nil
}

// enhancePixel enhances a single pixel for E-Ink display
func (e *EInkOptimizedEnhancer) enhancePixel(c color.Color) color.Color {
	r, g, b, a := c.RGBA()

	// Convert to 0-1 range
	rNorm := float64(r>>8) / 255.0
	gNorm := float64(g>>8) / 255.0
	bNorm := float64(b>>8) / 255.0

	// Apply contrast enhancement
	if e.contrastGain != 1.0 {
		rNorm = clampFloat((rNorm-0.5)*e.contrastGain + 0.5)
		gNorm = clampFloat((gNorm-0.5)*e.contrastGain + 0.5)
		bNorm = clampFloat((bNorm-0.5)*e.contrastGain + 0.5)
	}

	// Enhance color separation by boosting the dominant channel(s)
	if e.colorBoost > 0 {
		rNorm, gNorm, bNorm = e.boostColors(rNorm, gNorm, bNorm)
	}

	return color.RGBA{
		R: uint8(math.Round(rNorm * 255)),
		G: uint8(math.Round(gNorm * 255)),
		B: uint8(math.Round(bNorm * 255)),
		A: uint8(a >> 8),
	}
}

// boostColors enhances color separation by pushing colors toward the 6-color palette
func (e *EInkOptimizedEnhancer) boostColors(r, g, b float64) (float64, float64, float64) {
	// Find the closest palette color and blend toward it
	minDist := math.MaxFloat64
	var closest color.RGBA

	for _, pc := range EInk6ColorPalette {
		pr := float64(pc.R) / 255.0
		pg := float64(pc.G) / 255.0
		pb := float64(pc.B) / 255.0

		dist := (r-pr)*(r-pr) + (g-pg)*(g-pg) + (b-pb)*(b-pb)
		if dist < minDist {
			minDist = dist
			closest = pc
		}
	}

	// Blend toward the closest palette color
	pr := float64(closest.R) / 255.0
	pg := float64(closest.G) / 255.0
	pb := float64(closest.B) / 255.0

	r = r + (pr-r)*e.colorBoost*0.3
	g = g + (pg-g)*e.colorBoost*0.3
	b = b + (pb-b)*e.colorBoost*0.3

	return clampFloat(r), clampFloat(g), clampFloat(b)
}

// applySharpen applies a 3x3 sharpening kernel
func (e *EInkOptimizedEnhancer) applySharpen(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Sharpening kernel (unsharp mask style)
	// Center weight is boosted, edges are negative
	strength := e.sharpness

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get center pixel
			centerR, centerG, centerB, centerA := img.At(x, y).RGBA()

			// Calculate average of surrounding pixels
			var sumR, sumG, sumB float64
			count := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					nx, ny := x+dx, y+dy
					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						r, g, b, _ := img.At(nx, ny).RGBA()
						sumR += float64(r >> 8)
						sumG += float64(g >> 8)
						sumB += float64(b >> 8)
						count++
					}
				}
			}

			if count > 0 {
				avgR := sumR / float64(count)
				avgG := sumG / float64(count)
				avgB := sumB / float64(count)

				// Sharpen: enhance difference from average
				newR := float64(centerR>>8) + (float64(centerR>>8)-avgR)*strength
				newG := float64(centerG>>8) + (float64(centerG>>8)-avgG)*strength
				newB := float64(centerB>>8) + (float64(centerB>>8)-avgB)*strength

				result.Set(x, y, color.RGBA{
					R: uint8(clampFloat(newR/255.0) * 255),
					G: uint8(clampFloat(newG/255.0) * 255),
					B: uint8(clampFloat(newB/255.0) * 255),
					A: uint8(centerA >> 8),
				})
			} else {
				result.Set(x, y, img.At(x, y))
			}
		}
	}

	return result
}

func init() {
	// Register with moderate settings for E-Ink optimization
	RegisterEnhancer(NewEInkOptimizedEnhancer(0.5, 0.3, 1.1))
}
