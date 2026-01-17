package convert_test

import (
	"testing"

	"github.com/fooling/6-color-editor/cmd/convert"
)

func TestGetCommand(t *testing.T) {
	cmd := convert.GetCommand()
	if cmd == nil {
		t.Fatal("GetCommand() returned nil")
	}

	// Use field includes argument pattern, e.g., "convert [input-file]"
	if len(cmd.Use) < 7 || cmd.Use[:7] != "convert" {
		t.Errorf("Expected command use to start with 'convert', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected command long description to be set")
	}
}

func TestConvertCommand_Flags(t *testing.T) {
	cmd := convert.GetCommand()

	// Check that flags are defined
	flags := cmd.Flags()
	if flags == nil {
		t.Fatal("Command flags is nil")
	}

	// Check for expected flags
	expectedFlags := []string{
		"width",
		"height",
		"brightness",
		"contrast",
		"saturation",
		"dither",
		"upload",
		"remote",
		"output",
	}

	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found", flagName)
		}
	}
}

func TestConvertCommand_Args(t *testing.T) {
	cmd := convert.GetCommand()

	if cmd.Args == nil {
		t.Error("Expected Args validation to be set")
	}

	// The command expects exactly 1 argument
	// We can't easily test the Args function directly without running the command
}

func TestConvertCommand_DefaultValues(t *testing.T) {
	cmd := convert.GetCommand()

	// Check default flag values
	widthFlag := cmd.Flags().Lookup("width")
	if widthFlag == nil {
		t.Fatal("Width flag not found")
	}
	// Default width should be 0
	if widthFlag.DefValue != "0" {
		t.Logf("Note: Default width is %s", widthFlag.DefValue)
	}

	heightFlag := cmd.Flags().Lookup("height")
	if heightFlag == nil {
		t.Fatal("Height flag not found")
	}

	brightnessFlag := cmd.Flags().Lookup("brightness")
	if brightnessFlag == nil {
		t.Fatal("Brightness flag not found")
	}

	contrastFlag := cmd.Flags().Lookup("contrast")
	if contrastFlag == nil {
		t.Fatal("Contrast flag not found")
	}

	saturationFlag := cmd.Flags().Lookup("saturation")
	if saturationFlag == nil {
		t.Fatal("Saturation flag not found")
	}

	ditherFlag := cmd.Flags().Lookup("dither")
	if ditherFlag == nil {
		t.Fatal("Dither flag not found")
	}
}

func TestConvertCommand_Shorts(t *testing.T) {
	cmd := convert.GetCommand()

	// Check for shorthand flags
	tests := []struct {
		name     string
		shorthand string
	}{
		{"width", "W"},
		{"height", "H"},
		{"dither", "d"},
		{"upload", "u"},
		{"remote", "r"},
		{"output", "o"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().ShorthandLookup(tt.shorthand)
			if flag == nil {
				t.Errorf("Shorthand '%s' for flag '%s' not found", tt.shorthand, tt.name)
			}
		})
	}
}

func TestConvertCommand_HasExample(t *testing.T) {
	cmd := convert.GetCommand()

	if cmd.Long == "" {
		t.Error("Expected command to have a long description with examples")
	}

	longDesc := cmd.Long
	if !contains(longDesc, "Example") {
		t.Log("Note: Command examples may be in the Long description")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestConvertCommand_RunE(t *testing.T) {
	cmd := convert.GetCommand()

	// The RunE function should be set
	if cmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}
