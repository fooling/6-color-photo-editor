package main

import (
	"fmt"
	"os"

	"github.com/fooling/6-color-editor/cmd/root"
)

// version is overwritten at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	root.SetVersion(version)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
