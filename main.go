package main

import (
	"fmt"
	"os"

	"github.com/fooling/6-color-editor/cmd/root"
)

func main() {
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
