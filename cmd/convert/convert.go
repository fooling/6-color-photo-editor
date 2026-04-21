package convert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fooling/6-color-editor/internal/pipeline"
	"github.com/fooling/6-color-editor/pkg/uploader"
	"github.com/spf13/cobra"
	"image"
	"image/png"
)

var (
	width      int
	height     int
	brightness float64
	contrast   float64
	saturation float64
	dither     bool
	upload     bool
	remoteURL  string
	outputFile string
)

var convertCmd = &cobra.Command{
	Use:   "convert [input-file]",
	Short: "Convert an image to 6-color E-Ink format",
	Long: `Convert an image to the 6-color E-Ink palette with optional dithering.

The input is read from a file path, or from stdin when "-" is used as the
filename. The output is a PNG palette-quantized to the 6 display colors;
when --upload is given, the processed image is also encoded to BMP and
POSTed to the configured device URL.`,
	Example: `  # Basic file conversion to 800x480 PNG
  eink-6color convert photo.jpg -o output.png -W 800 -H 480

  # Read from stdin
  cat photo.jpg | eink-6color convert - -o output.png -W 800 -H 480

  # Convert and push to a device
  eink-6color convert photo.jpg -W 800 -H 480 \
    --upload --remote http://192.168.4.1/dataUP`,
	Args: cobra.ExactArgs(1),
	RunE: runConvert,
}

func init() {
	convertCmd.Flags().IntVarP(&width, "width", "W", 0, "Target width (0 = maintain aspect ratio)")
	convertCmd.Flags().IntVarP(&height, "height", "H", 0, "Target height (0 = maintain aspect ratio)")
	convertCmd.Flags().Float64Var(&brightness, "brightness", 0.0, "Brightness adjustment (-1.0 to 1.0)")
	convertCmd.Flags().Float64Var(&contrast, "contrast", 1.0, "Contrast adjustment (0.0 to 2.0+)")
	convertCmd.Flags().Float64Var(&saturation, "saturation", 1.0, "Saturation adjustment (0.0 to 2.0+)")
	convertCmd.Flags().BoolVarP(&dither, "dither", "d", true, "Enable Floyd-Steinberg dithering")
	convertCmd.Flags().BoolVarP(&upload, "upload", "u", false, "Upload to remote display after processing")
	convertCmd.Flags().StringVarP(&remoteURL, "remote", "r", "http://127.0.0.1/dataUP", "Remote display URL")
	convertCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (PNG format, omit for stdout)")
}

func runConvert(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Read input image
	var img image.Image
	var err error

	if inputFile == "-" {
		// Read from stdin
		img, err = decodeImage(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to decode image from stdin: %w", err)
		}
	} else {
		// Read from file
		file, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()

		img, err = decodeImage(file)
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}
	}

	// Configure processing
	config := &pipeline.ProcessConfig{
		Width:      width,
		Height:     height,
		Brightness: brightness,
		Contrast:   contrast,
		Saturation: saturation,
		Dither:     dither,
	}

	// Process image
	fmt.Fprintln(os.Stderr, "Processing image...")
	result := pipeline.ProcessWithConfig(img, config, false)

	// Output result
	var output io.Writer
	if outputFile == "" {
		output = os.Stdout
	} else {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		output = file
		fmt.Fprintf(os.Stderr, "Saved to %s\n", outputFile)
	}

	// Encode as PNG
	if err := png.Encode(output, result.Final); err != nil {
		return fmt.Errorf("failed to encode output: %w", err)
	}

	// Upload if requested
	if upload {
		fmt.Fprintln(os.Stderr, "Uploading to display...")
		upldr := uploader.NewUploader(&uploader.Config{
			RemoteURL: remoteURL,
			Timeout:   30 * time.Second,
		})

		if err := upldr.Upload(result.Final); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Upload complete!")
	}

	return nil
}

// decodeImage decodes an image from an io.Reader (supports JPEG and PNG)
func decodeImage(r io.Reader) (image.Image, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	// Log format to stderr
	fmt.Fprintf(os.Stderr, "Input format: %s, Size: %dx%d\n", format, img.Bounds().Dx(), img.Bounds().Dy())

	return img, nil
}

// GetCommand returns the convert command for registration
func GetCommand() *cobra.Command {
	return convertCmd
}
