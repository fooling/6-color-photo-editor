package uploader_test

import (
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

		if r.Header.Get("Content-Type") != "image/bmp" {
			t.Errorf("Expected Content-Type image/bmp, got %s", r.Header.Get("Content-Type"))
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

	if uploader.DefaultRemoteURL != "http://127.0.0.1:8080/esp/dataUP" {
		t.Errorf("DefaultRemoteURL = %s, want http://127.0.0.1:8080/esp/dataUP", uploader.DefaultRemoteURL)
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
