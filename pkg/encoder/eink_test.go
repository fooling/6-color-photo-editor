package encoder_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/core/palette"
	"github.com/fooling/6-color-editor/pkg/encoder"
)

func TestNewEInk(t *testing.T) {
	e := encoder.NewEInk()
	if e == nil {
		t.Fatal("NewEInk() returned nil")
	}
}

func TestEInk_Encode_ExactPaletteColors(t *testing.T) {
	e := encoder.NewEInk()

	tests := []struct {
		name  string
		color color.Color
		idx   palette.PaletteIndex
	}{
		{"black", palette.ColorBlack, palette.IndexBlack},
		{"white", palette.ColorWhite, palette.IndexWhite},
		{"green", palette.ColorGreen, palette.IndexGreen},
		{"blue", palette.ColorBlue, palette.IndexBlue},
		{"red", palette.ColorRed, palette.IndexRed},
		{"yellow", palette.ColorYellow, palette.IndexYellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a 1x1 image with the test color
			img := image.NewRGBA(image.Rect(0, 0, 1, 1))
			img.Set(0, 0, tt.color)

			data, err := e.Encode(img)
			if err != nil {
				t.Fatalf("Encode() returned error: %v", err)
			}

			// Verify header: width (2 bytes) + height (2 bytes)
			if len(data) != 5 { // 4 bytes header + 1 pixel
				t.Errorf("Expected 5 bytes, got %d", len(data))
			}

			// Verify pixel data is the correct index
			if data[4] != byte(tt.idx) {
				t.Errorf("Expected pixel index %d, got %d", tt.idx, data[4])
			}
		})
	}
}

func TestEInk_Encode_NonPaletteColor(t *testing.T) {
	e := encoder.NewEInk()

	// Create a 1x1 image with a color not in the palette
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{128, 128, 128, 255}) // Gray

	data, err := e.Encode(img)
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	// Should have encoded something (closest palette color)
	if len(data) != 5 {
		t.Errorf("Expected 5 bytes, got %d", len(data))
	}

	// The pixel index should be valid (0-5)
	idx := data[4]
	if idx > 5 {
		t.Errorf("Pixel index %d is out of range [0-5]", idx)
	}
}

func TestEInk_Encode_Header(t *testing.T) {
	e := encoder.NewEInk()

	tests := []struct {
		name            string
		width, height int
	}{
		{"1x1", 1, 1},
		{"10x10", 10, 10},
		{"296x296", 296, 296},
		{"100x50", 100, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, tt.width, tt.height))

			data, err := e.Encode(img)
			if err != nil {
				t.Fatalf("Encode() returned error: %v", err)
			}

			// Verify total size: 4 bytes header + width * height pixels
			expectedSize := 4 + tt.width*tt.height
			if len(data) != expectedSize {
				t.Errorf("Expected %d bytes, got %d", expectedSize, len(data))
			}

			// Verify width in header (big endian, first 2 bytes)
			width := int(data[0])<<8 | int(data[1])
			if width != tt.width {
				t.Errorf("Expected width %d in header, got %d", tt.width, width)
			}

			// Verify height in header (big endian, next 2 bytes)
			height := int(data[2])<<8 | int(data[3])
			if height != tt.height {
				t.Errorf("Expected height %d in header, got %d", tt.height, height)
			}
		})
	}
}

func TestEInk_Encode_AllPixels(t *testing.T) {
	e := encoder.NewEInk()

	// Create a 2x2 image with different colors
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, palette.ColorBlack)
	img.Set(1, 0, palette.ColorWhite)
	img.Set(0, 1, palette.ColorRed)
	img.Set(1, 1, palette.ColorBlue)

	data, err := e.Encode(img)
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	// Verify size: 4 bytes header + 4 pixels
	if len(data) != 8 {
		t.Errorf("Expected 8 bytes, got %d", len(data))
	}

	// Verify pixel data (row-major order)
	expected := []byte{
		byte(palette.IndexBlack),  // (0, 0)
		byte(palette.IndexWhite),  // (1, 0)
		byte(palette.IndexRed),    // (0, 1)
		byte(palette.IndexBlue),   // (1, 1)
	}

	for i := 0; i < 4; i++ {
		if data[4+i] != expected[i] {
			t.Errorf("Pixel %d: expected index %d, got %d", i, expected[i], data[4+i])
		}
	}
}

func TestEInk_Encode_EmptyImage(t *testing.T) {
	e := encoder.NewEInk()

	// Create a 0x0 image
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))

	data, err := e.Encode(img)
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	// Should only have header
	if len(data) != 4 {
		t.Errorf("Expected 4 bytes (header only), got %d", len(data))
	}
}

func TestEInk_Encode_LargeImage(t *testing.T) {
	e := encoder.NewEInk()

	// Create a larger image
	img := image.NewRGBA(image.Rect(0, 0, 296, 296))

	data, err := e.Encode(img)
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	// Verify size
	expectedSize := 4 + 296*296
	if len(data) != expectedSize {
		t.Errorf("Expected %d bytes, got %d", expectedSize, len(data))
	}
}

func TestEInk_Encode_ColorMatching(t *testing.T) {
	e := encoder.NewEInk()

	// Test that colors close to palette colors are mapped correctly
	tests := []struct {
		name  string
		color color.Color
		want  palette.PaletteIndex
	}{
		{
			name:  "dark gray -> black",
			color: color.RGBA{10, 10, 10, 255},
			want:  palette.IndexBlack,
		},
		{
			name:  "light gray -> white",
			color: color.RGBA{245, 245, 245, 255},
			want:  palette.IndexWhite,
		},
		{
			name:  "bright red",
			color: color.RGBA{255, 50, 50, 255},
			want:  palette.IndexRed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 1, 1))
			img.Set(0, 0, tt.color)

			data, err := e.Encode(img)
			if err != nil {
				t.Fatalf("Encode() returned error: %v", err)
			}

			idx := data[4]
			if idx != byte(tt.want) {
				t.Errorf("Expected index %d, got %d", tt.want, idx)
			}
		})
	}
}

func TestEInk_Encode_Rectangle(t *testing.T) {
	e := encoder.NewEInk()

	// Create a non-zero origin image
	img := image.NewRGBA(image.Rect(10, 20, 30, 40))

	data, err := e.Encode(img)
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	// Width should be 20 (30 - 10)
	width := int(data[0])<<8 | int(data[1])
	if width != 20 {
		t.Errorf("Expected width 20, got %d", width)
	}

	// Height should be 20 (40 - 20)
	height := int(data[2])<<8 | int(data[3])
	if height != 20 {
		t.Errorf("Expected height 20, got %d", height)
	}
}

func TestEInk_Encode_PreservesRowMajorOrder(t *testing.T) {
	e := encoder.NewEInk()

	// Create a 3x2 image
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			// Use white for all pixels
			img.Set(x, y, palette.ColorWhite)
		}
	}

	data, err := e.Encode(img)
	if err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}

	// All pixels should be white (index 1)
	for i := 0; i < 6; i++ {
		if data[4+i] != byte(palette.IndexWhite) {
			t.Errorf("Pixel %d: expected white (index 1), got %d", i, data[4+i])
		}
	}
}
