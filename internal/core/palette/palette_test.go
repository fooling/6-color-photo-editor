package palette_test

import (
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/core/palette"
)

func TestEInkColors(t *testing.T) {
	colors := palette.EInkColors()

	if len(colors) != palette.PaletteSize() {
		t.Errorf("Expected %d colors, got %d", palette.PaletteSize(), len(colors))
	}

	// Verify each color matches the constants
	expectedColors := []color.Color{
		palette.ColorBlack,
		palette.ColorWhite,
		palette.ColorGreen,
		palette.ColorBlue,
		palette.ColorRed,
		palette.ColorYellow,
	}

	for i, expected := range expectedColors {
		if colors[i] != expected {
			t.Errorf("Color at index %d does not match expected", i)
		}
	}
}

func TestPaletteSize(t *testing.T) {
	if palette.PaletteSize() != 6 {
		t.Errorf("Expected palette size 6, got %d", palette.PaletteSize())
	}
}

func TestIndexForColor(t *testing.T) {
	tests := []struct {
		name  string
		color color.Color
		want  int
	}{
		{"black", palette.ColorBlack, int(palette.IndexBlack)},
		{"white", palette.ColorWhite, int(palette.IndexWhite)},
		{"green", palette.ColorGreen, int(palette.IndexGreen)},
		{"blue", palette.ColorBlue, int(palette.IndexBlue)},
		{"red", palette.ColorRed, int(palette.IndexRed)},
		{"yellow", palette.ColorYellow, int(palette.IndexYellow)},
		{"unknown", color.RGBA{123, 123, 123, 255}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := palette.IndexForColor(tt.color)
			if got != tt.want {
				t.Errorf("IndexForColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorForIndex(t *testing.T) {
	tests := []struct {
		name string
		idx  palette.PaletteIndex
		want color.Color
	}{
		{"black", palette.IndexBlack, palette.ColorBlack},
		{"white", palette.IndexWhite, palette.ColorWhite},
		{"green", palette.IndexGreen, palette.ColorGreen},
		{"blue", palette.IndexBlue, palette.ColorBlue},
		{"red", palette.IndexRed, palette.ColorRed},
		{"yellow", palette.IndexYellow, palette.ColorYellow},
		{"invalid", 99, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := palette.ColorForIndex(tt.idx)
			if got != tt.want {
				t.Errorf("ColorForIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorConstants(t *testing.T) {
	// Test black
	r, g, b, a := palette.ColorBlack.RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0xffff {
		t.Errorf("ColorBlack has wrong values: r=%d, g=%d, b=%d, a=%d", r, g, b, a)
	}

	// Test white
	r, g, b, a = palette.ColorWhite.RGBA()
	if r != 0xffff || g != 0xffff || b != 0xffff || a != 0xffff {
		t.Errorf("ColorWhite has wrong values: r=%d, g=%d, b=%d, a=%d", r, g, b, a)
	}

	// Test green
	r, g, b, a = palette.ColorGreen.RGBA()
	if r != 0 || g != 0xffff || b != 0 || a != 0xffff {
		t.Errorf("ColorGreen has wrong values: r=%d, g=%d, b=%d, a=%d", r, g, b, a)
	}

	// Test blue
	r, g, b, a = palette.ColorBlue.RGBA()
	if r != 0 || g != 0 || b != 0xffff || a != 0xffff {
		t.Errorf("ColorBlue has wrong values: r=%d, g=%d, b=%d, a=%d", r, g, b, a)
	}

	// Test red
	r, g, b, a = palette.ColorRed.RGBA()
	if r != 0xffff || g != 0 || b != 0 || a != 0xffff {
		t.Errorf("ColorRed has wrong values: r=%d, g=%d, b=%d, a=%d", r, g, b, a)
	}

	// Test yellow
	r, g, b, a = palette.ColorYellow.RGBA()
	if r != 0xffff || g != 0xffff || b != 0 || a != 0xffff {
		t.Errorf("ColorYellow has wrong values: r=%d, g=%d, b=%d, a=%d", r, g, b, a)
	}
}
