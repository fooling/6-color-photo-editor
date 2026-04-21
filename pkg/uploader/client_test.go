package uploader_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fooling/6-color-editor/pkg/uploader"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		config *uploader.Config
	}{
		{"nil config", nil},
		{"empty config", &uploader.Config{}},
		{"with URL", &uploader.Config{RemoteURL: "http://example.com"}},
		{"with timeout", &uploader.Config{Timeout: 10 * time.Second}},
		{"full config", &uploader.Config{
			RemoteURL: "http://example.com",
			Timeout:   10 * time.Second,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := uploader.NewClient(tt.config)
			if c == nil {
				t.Fatal("NewClient() returned nil")
			}
		})
	}
}

func TestNewClient_DefaultValues(t *testing.T) {
	c := uploader.NewClient(nil)

	if c.RemoteURL() != uploader.DefaultRemoteURL {
		t.Errorf("Expected default remote URL %s, got %s", uploader.DefaultRemoteURL, c.RemoteURL())
	}
}

func TestNewClient_CustomURL(t *testing.T) {
	config := &uploader.Config{
		RemoteURL: "http://custom.example.com",
	}
	c := uploader.NewClient(config)

	if c.RemoteURL() != "http://custom.example.com" {
		t.Errorf("Expected custom remote URL, got %s", c.RemoteURL())
	}
}

func TestClient_Upload_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Errorf("Expected Content-Type application/octet-stream, got %s", r.Header.Get("Content-Type"))
		}

		body, _ := io.ReadAll(r.Body)
		// First byte is the STA network-mode flag (1) the firmware expects;
		// AP is 0, see pkg/uploader/client.go modeSTA doc.
		if len(body) < 1 || body[0] != 1 {
			t.Errorf("Expected first byte to be 1 (STA), got %d", body[0])
		}
		// Remaining bytes should be the payload we passed in, untouched.
		if want := []byte{1, 2, 3, 4, 5}; !bytes.Equal(body[1:], want) {
			t.Errorf("Expected payload %v after mode byte, got %v", want, body[1:])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &uploader.Config{
		RemoteURL: server.URL,
		Timeout:   5 * time.Second,
	}
	c := uploader.NewClient(config)

	data := []byte{1, 2, 3, 4, 5}
	err := c.Upload(data)
	if err != nil {
		t.Errorf("Upload() returned error: %v", err)
	}
}

func TestClient_Upload_Protocols(t *testing.T) {
	tests := []struct {
		name            string
		protocol        uploader.Protocol
		wantContentType string
		wantBodyFirst   []byte // expected prefix; full body = prefix + input data
		wantAcceptLang  bool   // legacy protocol must include Accept-Language
	}{
		{
			name:            "new STA prepends 0x01",
			protocol:        uploader.ProtocolNewSTA,
			wantContentType: "application/octet-stream",
			wantBodyFirst:   []byte{0x01},
		},
		{
			name:            "new AP prepends 0x00",
			protocol:        uploader.ProtocolNewAP,
			wantContentType: "application/octet-stream",
			wantBodyFirst:   []byte{0x00},
		},
		{
			name:            "legacy sends raw BMP with browser headers",
			protocol:        uploader.ProtocolLegacy,
			wantContentType: "image/bmp",
			wantBodyFirst:   nil,
			wantAcceptLang:  true,
		},
	}

	payload := []byte{0xAA, 0xBB, 0xCC}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotCT, gotAL string
			var gotBody []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotCT = r.Header.Get("Content-Type")
				gotAL = r.Header.Get("Accept-Language")
				gotBody, _ = io.ReadAll(r.Body)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			c := uploader.NewClient(&uploader.Config{
				RemoteURL: server.URL,
				Timeout:   5 * time.Second,
				Protocol:  tt.protocol,
			})
			if err := c.Upload(payload); err != nil {
				t.Fatalf("Upload returned %v", err)
			}

			if gotCT != tt.wantContentType {
				t.Errorf("Content-Type = %q, want %q", gotCT, tt.wantContentType)
			}

			wantBody := append(append([]byte{}, tt.wantBodyFirst...), payload...)
			if !bytes.Equal(gotBody, wantBody) {
				t.Errorf("Body = %v, want %v", gotBody, wantBody)
			}

			if tt.wantAcceptLang && gotAL == "" {
				t.Errorf("Expected Accept-Language on legacy protocol, got empty")
			}
			if !tt.wantAcceptLang && gotAL != "" {
				t.Errorf("Unexpected Accept-Language on new protocol: %q", gotAL)
			}
		})
	}
}

func TestClient_Upload_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	config := &uploader.Config{
		RemoteURL: server.URL,
		Timeout:   5 * time.Second,
	}
	c := uploader.NewClient(config)

	data := []byte{1, 2, 3, 4, 5}
	err := c.Upload(data)
	if err == nil {
		t.Error("Expected error for 500 response, got nil")
	}
}

func TestClient_Upload_NetworkError(t *testing.T) {
	config := &uploader.Config{
		RemoteURL: "http://invalid.local:9999",
		Timeout:   100 * time.Millisecond,
	}
	c := uploader.NewClient(config)

	data := []byte{1, 2, 3, 4, 5}
	err := c.Upload(data)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestClient_Upload_EmptyData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &uploader.Config{
		RemoteURL: server.URL,
		Timeout:   5 * time.Second,
	}
	c := uploader.NewClient(config)

	data := []byte{}
	err := c.Upload(data)
	if err != nil {
		t.Errorf("Upload() with empty data returned error: %v", err)
	}
}

func TestClient_Upload_LargeData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &uploader.Config{
		RemoteURL: server.URL,
		Timeout:   5 * time.Second,
	}
	c := uploader.NewClient(config)

	// Create a 100KB payload
	data := make([]byte, 100*1024)
	err := c.Upload(data)
	if err != nil {
		t.Errorf("Upload() with large data returned error: %v", err)
	}
}

func TestClient_Upload_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &uploader.Config{
		RemoteURL: server.URL,
		Timeout:   50 * time.Millisecond,
	}
	c := uploader.NewClient(config)

	data := []byte{1, 2, 3}
	err := c.Upload(data)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestClient_RemoteURL(t *testing.T) {
	config := &uploader.Config{
		RemoteURL: "http://test.example.com",
	}
	c := uploader.NewClient(config)

	if c.RemoteURL() != "http://test.example.com" {
		t.Errorf("RemoteURL() returned unexpected value: %s", c.RemoteURL())
	}
}

func TestDefaultRemoteURL(t *testing.T) {
	if uploader.DefaultRemoteURL == "" {
		t.Error("DefaultRemoteURL is empty")
	}

	if uploader.DefaultRemoteURL != "http://127.0.0.1/dataUP" {
		t.Errorf("DefaultRemoteURL = %s, want http://127.0.0.1/dataUP", uploader.DefaultRemoteURL)
	}
}

func TestUploadTimeout(t *testing.T) {
	if uploader.UploadTimeout != 30*time.Second {
		t.Errorf("UploadTimeout = %v, want 30s", uploader.UploadTimeout)
	}
}

func TestClient_Upload_StatusCodeChecks(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		shouldFail bool
	}{
		{"200 OK", http.StatusOK, false},
		{"201 Created", http.StatusCreated, true},  // Only 200 is accepted
		{"204 No Content", http.StatusNoContent, true}, // Only 200 is accepted
		{"400 Bad Request", http.StatusBadRequest, true},
		{"404 Not Found", http.StatusNotFound, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			config := &uploader.Config{
				RemoteURL: server.URL,
				Timeout:   5 * time.Second,
			}
			c := uploader.NewClient(config)

			data := []byte{1, 2, 3}
			err := c.Upload(data)

			if tt.shouldFail && err == nil {
				t.Errorf("Expected error for status %d, got nil", tt.statusCode)
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected no error for status %d, got %v", tt.statusCode, err)
			}
		})
	}
}
