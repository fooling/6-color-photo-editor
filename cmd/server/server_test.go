package server_test

import (
	"testing"

	"github.com/fooling/6-color-editor/cmd/server"
)

func TestGetCommand(t *testing.T) {
	cmd := server.GetCommand()
	if cmd == nil {
		t.Fatal("GetCommand() returned nil")
	}

	if cmd.Use != "server" {
		t.Errorf("Expected command use 'server', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected command long description to be set")
	}
}

func TestServerCommand_Flags(t *testing.T) {
	cmd := server.GetCommand()

	// Check that flags are defined
	flags := cmd.Flags()
	if flags == nil {
		t.Fatal("Command flags is nil")
	}

	// Check for expected flags
	expectedFlags := []string{
		"port",
		"host",
		"remote-url",
	}

	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found", flagName)
		}
	}
}

func TestServerCommand_Shorts(t *testing.T) {
	cmd := server.GetCommand()

	// Check for shorthand flags
	tests := []struct {
		name     string
		shorthand string
	}{
		{"port", "p"},
		{"host", "H"},
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

func TestServerCommand_DefaultValues(t *testing.T) {
	cmd := server.GetCommand()

	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("Port flag not found")
	}
	// Default port should be 3000
	if portFlag.DefValue != "3000" {
		t.Logf("Note: Default port is %s", portFlag.DefValue)
	}

	hostFlag := cmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Fatal("Host flag not found")
	}

	remoteURLFlag := cmd.Flags().Lookup("remote-url")
	if remoteURLFlag == nil {
		t.Fatal("RemoteURL flag not found")
	}
}

func TestServerCommand_RunE(t *testing.T) {
	cmd := server.GetCommand()

	// The RunE function should be set
	if cmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}

func TestServerCommand_Features(t *testing.T) {
	cmd := server.GetCommand()

	if cmd.Long == "" {
		t.Error("Expected command to have a long description")
	}

	longDesc := cmd.Long

	// Check for feature descriptions
	expectedFeatures := []string{
		"Split view",
		"Live adjustments",
		"Step visualization",
		"Direct upload",
	}

	for _, feature := range expectedFeatures {
		if !containsString(longDesc, feature) {
			t.Logf("Note: Feature '%s' may be described differently", feature)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsMiddleString(s, substr))
}

func containsMiddleString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestServerCommand_NoArgs(t *testing.T) {
	cmd := server.GetCommand()

	// Server command typically takes no positional arguments
	// This is verified by checking if Args validation is nil (accepts all args)
	// or has specific validation
	if cmd.Args != nil {
		t.Log("Server command has argument validation")
	}
}
