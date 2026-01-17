package root

import (
	"github.com/fooling/6-color-editor/cmd/convert"
	"github.com/fooling/6-color-editor/cmd/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "eink-convert",
	Short: "E-Ink 6-Color Image Processing Tool",
	Long: `A robust, cross-platform, single-binary image processing tool for E-Ink displays.

Supports conversion to 6-color palette (Black, White, Red, Green, Blue, Yellow)
with Floyd-Steinberg dithering. Can run as CLI tool or Web Server.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(convert.GetCommand())
	rootCmd.AddCommand(server.GetCommand())
}
