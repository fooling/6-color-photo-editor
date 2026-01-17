// Package uploader provides functionality to upload images to E-Ink displays.
package uploader

import (
	"image"

	"github.com/fooling/6-color-editor/pkg/encoder"
)

// Uploader handles encoding and uploading images to E-Ink displays.
// Deprecated: Use encoder.EInk for encoding and Client for uploading separately.
type Uploader struct {
	client  *Client
	encoder *encoder.EInk
}

// NewUploader creates a new uploader with the given configuration.
func NewUploader(config *Config) *Uploader {
	return &Uploader{
		client:  NewClient(config),
		encoder: encoder.NewEInk(),
	}
}

// Upload encodes and uploads an image to the configured remote endpoint.
func (u *Uploader) Upload(img image.Image) error {
	data, err := u.encoder.Encode(img)
	if err != nil {
		return err
	}
	return u.client.Upload(data)
}

// UploadBytes uploads raw binary data to the remote endpoint.
func (u *Uploader) UploadBytes(data []byte) error {
	return u.client.Upload(data)
}
