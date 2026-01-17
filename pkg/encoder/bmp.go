// Package encoder provides encoding functionality for E-Ink display formats.
//
// BMP Encoder:
//   - Supports 24-bit RGB format
//   - Row padding to 4-byte boundaries (BMP requirement)
//   - Little-endian byte order
//
// Common E-Ink resolutions:
//   - 800 x 480 (landscape)
//   - 480 x 800 (portrait)
package encoder

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"io"
)

const (
	bmpHeaderSize  = 14
	bmpInfoSize    = 40 // BITMAPINFOHEADER size
	bmpPlanes      = 1
	bmpBits        = 24 // 24-bit RGB
	bmpCompression = 0  // BI_RGB
)

// BMP encodes images to BMP format.
type BMP struct{}

// NewBMP creates a new BMP encoder.
func NewBMP() *BMP {
	return &BMP{}
}

// Encode converts an image to BMP format (24-bit RGB).
//
// BMP Format:
//   - File Header (14 bytes): signature, file size, offset to pixel data
//   - Info Header (40 bytes): dimensions, color depth, etc.
//   - Pixel Data: bottom-to-top, BGR order, row padding to 4-byte boundary
//
// Example:
//   b := encoder.NewBMP()
//   data, err := b.Encode(img)
func (b *BMP) Encode(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate row size with padding to 4-byte boundary
	rowSize := ((width*3 + 3) / 4) * 4
	paddingSize := rowSize - (width * 3)

	// Calculate file sizes
	pixelDataSize := rowSize * height
	fileSize := bmpHeaderSize + bmpInfoSize + pixelDataSize
	offsetToPixels := bmpHeaderSize + bmpInfoSize

	buf := new(bytes.Buffer)

	// Write file header
	if err := writeFileHeader(buf, fileSize, offsetToPixels); err != nil {
		return nil, err
	}

	// Write info header
	if err := writeInfoHeader(buf, width, height, pixelDataSize); err != nil {
		return nil, err
	}

	// Write pixel data (bottom-to-top, BGR)
	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert 16-bit to 8-bit
			bgr := []byte{
				byte(b >> 8),
				byte(g >> 8),
				byte(r >> 8),
			}
			if _, err := buf.Write(bgr); err != nil {
				return nil, err
			}
		}
		// Write padding
		for i := 0; i < paddingSize; i++ {
			buf.WriteByte(0)
		}
	}

	return buf.Bytes(), nil
}

func writeFileHeader(w io.Writer, fileSize, offsetToPixels int) error {
	header := struct {
		Signature [2]byte // 'BM'
		FileSize  uint32
		Reserved  uint16
		Reserved2 uint16
		Offset    uint32
	}{
		Signature: [2]byte{'B', 'M'},
		FileSize:  uint32(fileSize),
		Offset:    uint32(offsetToPixels),
	}
	return binary.Write(w, binary.LittleEndian, header)
}

func writeInfoHeader(w io.Writer, width, height, pixelDataSize int) error {
	header := struct {
		Size          uint32 // 40 for BITMAPINFOHEADER
		Width         int32
		Height        int32
		Planes        uint16
		BitsPerPixel  uint16
		Compression   uint32
		ImageSize     uint32
		XPixelsPerM   int32
		YPixelsPerM   int32
		ColorsUsed    uint32
		ColorsImportant uint32
	}{
		Size:            bmpInfoSize,
		Width:           int32(width),
		Height:          int32(height),
		Planes:          bmpPlanes,
		BitsPerPixel:    bmpBits,
		Compression:     bmpCompression,
		ImageSize:       uint32(pixelDataSize),
		XPixelsPerM:     2835, // 72 DPI
		YPixelsPerM:     2835, // 72 DPI
		ColorsUsed:      0,
		ColorsImportant: 0,
	}
	return binary.Write(w, binary.LittleEndian, header)
}

// bgrImage wraps an image.Image to provide BGR color access
type bgrImage struct {
	image.Image
}

// At returns the color in BGR order for BMP encoding
func (img *bgrImage) At(x, y int) color.Color {
	c := img.Image.At(x, y)
	r, g, b, a := c.RGBA()
	// Convert to 8-bit BGR
	return color.RGBA{
		R: uint8(b >> 8),
		G: uint8(g >> 8),
		B: uint8(r >> 8),
		A: uint8(a >> 8),
	}
}
