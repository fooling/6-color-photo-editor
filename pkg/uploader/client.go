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
	DefaultRemoteURL = "http://127.0.0.1/dataUP"
	// UploadTimeout is the default timeout for uploads.
	UploadTimeout = 30 * time.Second
)

// Network-mode bytes for the new-firmware protocol. Verified against the
// device's script.min.js (STA=1, AP=0).
const (
	modeAP  byte = 0
	modeSTA byte = 1
)

// Protocol selects which wire format Upload uses.
type Protocol int

const (
	// ProtocolNewSTA is the new-firmware protocol with the STA network-mode
	// byte prepended. Body: [0x01][BMP]. Content-Type: application/octet-stream.
	// This is the default (zero value).
	ProtocolNewSTA Protocol = iota
	// ProtocolNewAP is the new-firmware protocol with the AP network-mode
	// byte prepended. Body: [0x00][BMP]. Content-Type: application/octet-stream.
	ProtocolNewAP
	// ProtocolLegacy is the pre-1.2.0 firmware protocol (matches repo tag
	// v0.0.1). Body: raw BMP, no mode byte. Content-Type: image/bmp, with
	// browser-ish Accept / Accept-Language / Accept-Encoding headers.
	ProtocolLegacy
)

// Client handles HTTP uploads to E-Ink displays.
// It manages HTTP connection pooling and timeouts for efficient uploads.
type Client struct {
	client    *http.Client
	remoteURL string
	protocol  Protocol
}

// Config holds client configuration.
type Config struct {
	// RemoteURL is the endpoint URL for uploading images.
	// If empty, DefaultRemoteURL is used.
	RemoteURL string
	// Timeout is the HTTP request timeout.
	// If zero, UploadTimeout is used.
	Timeout time.Duration
	// Protocol selects the wire format. Zero value = ProtocolNewSTA.
	Protocol Protocol
}

// NewClient creates a new HTTP client with the given configuration.
// The client reuses connections for multiple uploads.
//
// Example:
//   client := uploader.NewClient(&uploader.Config{
//       RemoteURL: "http://192.168.1.100/dataUP",
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
		protocol:  config.Protocol,
	}
}

// Upload sends a payload to the configured remote endpoint using the
// selected Protocol. For the new-firmware protocols (STA/AP) a network-mode
// byte is prepended to the BMP data; for ProtocolLegacy the BMP is sent
// as-is with image/bmp plus the v0.0.1-era browser-ish headers.
//
// Returns an error if the upload fails or the server returns a non-200 status.
//
// Example:
//   data, _ := encoder.NewEInk().Encode(img)
//   err := client.Upload(data)
func (c *Client) Upload(data []byte) error {
	reqBody, contentType, extraHeaders, logLine := c.buildRequest(data)

	log.Printf("[Uploader] ========================================")
	log.Printf("[Uploader] HTTP Request Details:")
	log.Printf("[Uploader]   Method: POST")
	log.Printf("[Uploader]   URL: %s", c.remoteURL)
	log.Printf("[Uploader]   Protocol: %s", logLine)
	log.Printf("[Uploader]   Content-Length: %d bytes", len(reqBody))
	log.Printf("[Uploader] Request Headers:")
	log.Printf("[Uploader]   Content-Type: %s", contentType)
	for k, v := range extraHeaders {
		log.Printf("[Uploader]   %s: %s", k, v)
	}

	req, err := http.NewRequest("POST", c.remoteURL, bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("[Uploader] Failed to create request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

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

// buildRequest assembles the body, content type and extra headers according
// to the client's protocol. The returned logLine is a human-readable tag
// for diagnostics only.
func (c *Client) buildRequest(bmp []byte) (reqBody []byte, contentType string, extraHeaders map[string]string, logLine string) {
	switch c.protocol {
	case ProtocolNewAP:
		return prependMode(modeAP, bmp), "application/octet-stream", nil, "new/AP (mode=0x00)"
	case ProtocolLegacy:
		// v0.0.1 wire format: raw BMP + image/bmp + browser-ish Accept* headers.
		return bmp, "image/bmp", map[string]string{
			"Accept":          "*/*",
			"Accept-Encoding": "gzip, deflate",
			"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		}, "legacy (pre-1.2.0 firmware)"
	case ProtocolNewSTA:
		fallthrough
	default:
		return prependMode(modeSTA, bmp), "application/octet-stream", nil, "new/STA (mode=0x01)"
	}
}

func prependMode(mode byte, bmp []byte) []byte {
	out := make([]byte, 1+len(bmp))
	out[0] = mode
	copy(out[1:], bmp)
	return out
}
