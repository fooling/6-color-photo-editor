// Package uploader provides functionality to upload images to E-Ink displays.
//
// The package includes an HTTP client for uploading binary encoded images
// to E-Ink display endpoints.
package uploader

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	// DefaultRemoteURL is the default ESP32 endpoint.
	DefaultRemoteURL = "http://127.0.0.1:8080/esp/dataUP"
	// UploadTimeout is the default timeout for uploads.
	UploadTimeout = 30 * time.Second
)

// Client handles HTTP uploads to E-Ink displays.
// It manages HTTP connection pooling and timeouts for efficient uploads.
type Client struct {
	client    *http.Client
	remoteURL string
}

// Config holds client configuration.
type Config struct {
	// RemoteURL is the endpoint URL for uploading images.
	// If empty, DefaultRemoteURL is used.
	RemoteURL string
	// Timeout is the HTTP request timeout.
	// If zero, UploadTimeout is used.
	Timeout time.Duration
}

// NewClient creates a new HTTP client with the given configuration.
// The client reuses connections for multiple uploads.
//
// Example:
//   client := uploader.NewClient(&uploader.Config{
//       RemoteURL: "http://192.168.1.100:8080/upload",
//       Timeout: 10 * time.Second,
//   })
func NewClient(config *Config) *Client {
	if config == nil {
		config = &Config{}
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = UploadTimeout
	}

	remoteURL := config.RemoteURL
	if remoteURL == "" {
		remoteURL = DefaultRemoteURL
	}

	return &Client{
		client: &http.Client{
			Timeout: timeout,
		},
		remoteURL: remoteURL,
	}
}

// Upload sends binary data to the configured remote endpoint.
// The data should be in the E-Ink binary format (width, height, pixel data).
// The request uses POST method with "application/octet-stream" content type.
//
// Returns an error if the upload fails or the server returns a non-200 status.
//
// Example:
//   data, _ := encoder.NewEInk().Encode(img)
//   err := client.Upload(data)
func (c *Client) Upload(data []byte) error {
	log.Printf("[Uploader] Starting upload to: %s", c.remoteURL)
	log.Printf("[Uploader] Data size: %d bytes", len(data))

	req, err := http.NewRequest("POST", c.remoteURL, bytes.NewReader(data))
	if err != nil {
		log.Printf("[Uploader] Failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	log.Printf("[Uploader] Sending request...")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("[Uploader] Request failed: %v", err)
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[Uploader] Response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Uploader] Error response body: %s", string(body))
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("[Uploader] Upload successful!")
	return nil
}

// RemoteURL returns the configured remote URL.
func (c *Client) RemoteURL() string {
	return c.remoteURL
}
