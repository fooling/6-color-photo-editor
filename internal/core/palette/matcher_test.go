package palette_test

import (
	"image/color"
	"testing"

	"github.com/fooling/6-color-editor/internal/core/palette"
)

func TestNewMatcher(t *testing.T) {
	m := palette.NewMatcher()
	if m == nil {
		t.Fatal("NewMatcher() returned nil")
	}
}

func TestMatcher_FindClosestColor_ExactMatch(t *testing.T) {
	m := palette.NewMatcher()

	tests := []struct {
		name  string
		color color.Color
		want  color.Color
	}{
		{"black exact", color.Black, palette.ColorBlack},
		{"white exact", color.White, palette.ColorWhite},
		{"green exact", palette.ColorGreen, palette.ColorGreen},
		{"blue exact", palette.ColorBlue, palette.ColorBlue},
		{"red exact", palette.ColorRed, palette.ColorRed},
		{"yellow exact", palette.ColorYellow, palette.ColorYellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.FindClosestColor(tt.color)
			if got != tt.want {
				t.Errorf("FindClosestColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatcher_FindClosestColor_ClosestMatch(t *testing.T) {
	m := palette.NewMatcher()

	tests := []struct {
		name  string
		color color.Color
		want  color.Color
	}{
		{
			name:  "dark gray close to black",
			color: color.RGBA{10, 10, 10, 255},
			want:  palette.ColorBlack,
		},
		{
			name:  "light gray close to white",
			color: color.RGBA{245, 245, 245, 255},
			want:  palette.ColorWhite,
		},
		{
			name:  "cyan close to blue",
			color: color.RGBA{0, 100, 200, 255},
			want:  palette.ColorBlue,
		},
		{
			name:  "lime close to green",
			color: color.RGBA{50, 255, 50, 255},
			want:  palette.ColorGreen,
		},
		{
			name:  "orange close to red",
			color: color.RGBA{255, 100, 0, 255},
			want:  palette.ColorRed,
		},
		{
			name:  "bright yellow",
			color: color.RGBA{255, 255, 50, 255},
			want:  palette.ColorYellow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.FindClosestColor(tt.color)
			if got != tt.want {
				t.Errorf("FindClosestColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatcher_FindClosestColor_EdgeCases(t *testing.T) {
	m := palette.NewMatcher()

	// Test pure RGB colors
	tests := []struct {
		name  string
		color color.Color
		want  color.Color
	}{
		{
			name:  "pure red",
			color: color.RGBA{255, 0, 0, 255},
			want:  palette.ColorRed,
		},
		{
			name:  "pure green",
			color: color.RGBA{0, 255, 0, 255},
			want:  palette.ColorGreen,
		},
		{
			name:  "pure blue",
			color: color.RGBA{0, 0, 255, 255},
			want:  palette.ColorBlue,
		},
		{
			name:  "magenta (equidistant from White, Red, Blue - White is first)",
			color: color.RGBA{255, 0, 255, 255},
			want:  palette.ColorWhite, // White comes first in palette
		},
		{
			name:  "gray (middle of black and white)",
			color: color.RGBA{128, 128, 128, 255},
			want:  palette.ColorWhite, // closer to white than black
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.FindClosestColor(tt.color)
			if got != tt.want {
				t.Errorf("FindClosestColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatcher_FindClosestIndex(t *testing.T) {
	m := palette.NewMatcher()

	tests := []struct {
		name  string
		color color.Color
		want  palette.PaletteIndex
	}{
		{
			name:  "black",
			color: color.Black,
			want:  palette.IndexBlack,
		},
		{
			name:  "white",
			color: color.White,
			want:  palette.IndexWhite,
		},
		{
			name:  "green",
			color: palette.ColorGreen,
			want:  palette.IndexGreen,
		},
		{
			name:  "blue",
			color: palette.ColorBlue,
			want:  palette.IndexBlue,
		},
		{
			name:  "red",
			color: palette.ColorRed,
			want:  palette.IndexRed,
		},
		{
			name:  "yellow",
			color: palette.ColorYellow,
			want:  palette.IndexYellow,
		},
		{
			name:  "dark gray should match black",
			color: color.RGBA{10, 10, 10, 255},
			want:  palette.IndexBlack,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.FindClosestIndex(tt.color)
			if got != tt.want {
				t.Errorf("FindClosestIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatcher_FindClosestColor_EuclideanDistance(t *testing.T) {
	m := palette.NewMatcher()

	// Test that the matcher finds the truly closest color
	// Red: (255, 0, 0), Yellow: (255, 255, 0)
	// Orange: (255, 128, 0) is closer to yellow than red
	// Distance to red: 0^2 + 128^2 + 0^2 = 16384
	// Distance to yellow: 0^2 + 127^2 + 0^2 = 16129
	orange := color.RGBA{255, 128, 0, 255}

	// Calculate distances manually to verify
	closest := m.FindClosestColor(orange)

	// Verify it's yellow (should be closer than red)
	r, g, b, _ := closest.RGBA()
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

	if r8 == 255 && g8 == 255 && b8 == 0 {
		// Correctly identified as yellow
	} else if r8 == 255 && g8 == 0 && b8 == 0 {
		// Got red instead - yellow is closer
		t.Error("Orange should be closer to yellow than red, but got red")
	} else {
		t.Errorf("Unexpected closest color for orange: R=%d, G=%d, B=%d", r8, g8, b8)
	}
}
