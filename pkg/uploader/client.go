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
	log.Printf("[Uploader] ========================================")
	log.Printf("[Uploader] HTTP Request Details:")
	log.Printf("[Uploader]   Method: POST")
	log.Printf("[Uploader]   URL: %s", c.remoteURL)
	log.Printf("[Uploader]   Content-Length: %d bytes", len(data))
	log.Printf("[Uploader] Request Headers:")
	log.Printf("[Uploader]   Content-Type: image/bmp")
	log.Printf("[Uploader]   Accept: */*")
	log.Printf("[Uploader]   Accept-Encoding: gzip, deflate")
	log.Printf("[Uploader]   Accept-Language: zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")

	req, err := http.NewRequest("POST", c.remoteURL, bytes.NewReader(data))
	if err != nil {
		log.Printf("[Uploader] Failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "image/bmp")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")

	log.Printf("[Uploader] ----------------------------------------")
	log.Printf("[Uploader] Sending request...")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("[Uploader] Request failed: %v", err)
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[Uploader] Response Status:")
	log.Printf("[Uploader]   Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("[Uploader] Response Headers:")
	for key, values := range resp.Header {
		for _, value := range values {
			log.Printf("[Uploader]   %s: %s", key, value)
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Uploader] Error Response Body (%d bytes):", len(body))
		log.Printf("[Uploader]   %s", string(body))
		log.Printf("[Uploader] ========================================")
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[Uploader] Response Body (%d bytes):", len(body))
	if len(body) > 0 {
		log.Printf("[Uploader]   %s", string(body))
	} else {
		log.Printf("[Uploader]   (empty)")
	}
	log.Printf("[Uploader] Upload successful!")
	log.Printf("[Uploader] ========================================")
	return nil
}

// RemoteURL returns the configured remote URL.
func (c *Client) RemoteURL() string {
	return c.remoteURL
}
