package root

import (
	"github.com/fooling/6-color-editor/cmd/convert"
	"github.com/fooling/6-color-editor/cmd/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "eink-6color",
	Short: "E-Ink 6-color image processing tool",
	Long: `A cross-platform, single-binary image processor for 6-color E-Ink displays.

Supports conversion to the 6-color palette (Black, White, Red, Green, Blue, Yellow)
with Floyd-Steinberg dithering. Runs as either a CLI (convert) or a web server (server).`,
	Example: `  # Launch the web UI on the default port (3000)
  eink-6color server

  # Launch on a custom port
  eink-6color server --port 8080

  # One-shot convert to 800x480 PNG
  eink-6color convert photo.jpg -o out.png -W 800 -H 480

  # Convert and push to a device in one step
  eink-6color convert photo.jpg -W 800 -H 480 \
    --upload --remote http://192.168.4.1/dataUP`,
}

// SetVersion wires the build-time version string into --version output.
func SetVersion(v string) {
	rootCmd.Version = v
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(convert.GetCommand())
	rootCmd.AddCommand(server.GetCommand())
}
