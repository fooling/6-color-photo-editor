// Package uploader provides functionality to upload images to E-Ink displays.
package uploader

import (
	"bytes"
	"encoding/binary"
	"image"
)

// Uploader handles encoding and uploading images to E-Ink displays.
// Deprecated: Use encoder.EInk for encoding and Client for uploading separately.
type Uploader struct {
	client *Client
}

// NewUploader creates a new uploader with the given configuration.
func NewUploader(config *Config) *Uploader {
	return &Uploader{
		client: NewClient(config),
	}
}

// Upload encodes image to BMP format and uploads to the configured remote endpoint.
func (u *Uploader) Upload(img image.Image) error {
	data, err := encodeBMP(img)
	if err != nil {
		return err
	}
	return u.client.Upload(data)
}

// encodeBMP encodes an image to BMP format (24-bit RGB).
func encodeBMP(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate row size with padding to 4-byte boundary
	rowSize := ((width*3 + 3) / 4) * 4
	paddingSize := rowSize - (width * 3)

	// Calculate file sizes
	const (
		bmpHeaderSize = 14
		bmpInfoSize   = 40
	)
	pixelDataSize := rowSize * height
	fileSize := bmpHeaderSize + bmpInfoSize + pixelDataSize
	offsetToPixels := bmpHeaderSize + bmpInfoSize

	buf := new(bytes.Buffer)

	// Write file header
	header := struct {
		Signature [2]byte
		FileSize  uint32
		Reserved  uint16
		Reserved2 uint16
		Offset    uint32
	}{
		Signature: [2]byte{'B', 'M'},
		FileSize:  uint32(fileSize),
		Offset:    uint32(offsetToPixels),
	}
	binary.Write(buf, binary.LittleEndian, header)

	// Write info header
	info := struct {
		Size            uint32
		Width           int32
		Height          int32
		Planes          uint16
		BitsPerPixel    uint16
		Compression     uint32
		ImageSize       uint32
		XPixelsPerM     int32
		YPixelsPerM     int32
		ColorsUsed      uint32
		ColorsImportant uint32
	}{
		Size:            bmpInfoSize,
		Width:           int32(width),
		Height:          int32(height),
		Planes:          1,
		BitsPerPixel:    24,
		Compression:     0,
		ImageSize:       uint32(pixelDataSize),
		XPixelsPerM:     2835,
		YPixelsPerM:     2835,
		ColorsUsed:      0,
		ColorsImportant: 0,
	}
	binary.Write(buf, binary.LittleEndian, info)

	// Write pixel data (bottom-to-top, BGR)
	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to 8-bit BGR
			buf.WriteByte(byte(b >> 8))
			buf.WriteByte(byte(g >> 8))
			buf.WriteByte(byte(r >> 8))
		}
		// Write padding
		for i := 0; i < paddingSize; i++ {
			buf.WriteByte(0)
		}
	}

	return buf.Bytes(), nil
}

// UploadBytes uploads raw binary data to the remote endpoint.
func (u *Uploader) UploadBytes(data []byte) error {
	return u.client.Upload(data)
}
