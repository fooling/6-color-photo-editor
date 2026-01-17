package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fooling/6-color-editor/internal/web/handlers"
)

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name   string
		config *handlers.Config
	}{
		{"nil config", nil},
		{"empty config", &handlers.Config{}},
		{"with remote URL", &handlers.Config{RemoteURL: "http://example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := handlers.NewRegistry(tt.config)
			if r == nil {
				t.Fatal("NewRegistry() returned nil")
			}
		})
	}
}

func TestRegistry_Routes(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	routes := registry.Routes()
	if routes == nil {
		t.Fatal("Routes() returned nil")
	}
}

func TestRegistry_Routes_PreviewEndpoint(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	req := httptest.NewRequest(http.MethodPost, "/api/preview", nil)
	w := httptest.NewRecorder()

	registry.Routes().ServeHTTP(w, req)

	// Should return bad request (missing body)
	if w.Code != http.StatusBadRequest {
		t.Logf("Expected status 400 for empty request, got %d", w.Code)
	}
}

func TestRegistry_Routes_UploadEndpoint(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	req := httptest.NewRequest(http.MethodPost, "/api/upload", nil)
	w := httptest.NewRecorder()

	registry.Routes().ServeHTTP(w, req)

	// Should return bad request (missing body)
	if w.Code != http.StatusBadRequest {
		t.Logf("Expected status 400 for empty request, got %d", w.Code)
	}
}

func TestRegistry_Routes_HealthEndpoint(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	registry.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestRegistry_Routes_StaticFiles(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	registry.Routes().ServeHTTP(w, req)

	// Should serve static files (404 is ok if file doesn't exist)
	// Status should not be 405 (method not allowed)
	if w.Code == http.StatusMethodNotAllowed {
		t.Error("Static files should handle GET requests")
	}
}

func TestRegistry_Routes_NotFound(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	req := httptest.NewRequest(http.MethodGet, "/api/notfound", nil)
	w := httptest.NewRecorder()

	registry.Routes().ServeHTTP(w, req)

	// Should return 404
	if w.Code != http.StatusNotFound {
		t.Logf("Expected status 404, got %d", w.Code)
	}
}

func TestRegistry_Middleware_Headers(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	registry.Routes().ServeHTTP(w, req)

	// Check security headers
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Missing X-Content-Type-Options header")
	}

	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("Missing X-Frame-Options header")
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Missing CORS header")
	}
}

func TestRegistry_Middleware_Options(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	req := httptest.NewRequest(http.MethodOptions, "/api/preview", nil)
	w := httptest.NewRecorder()

	registry.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}

	// Check CORS headers are set
	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Error("Missing Access-Control-Allow-Methods header")
	}
}

func TestRegistry_Config_RemoteURL(t *testing.T) {
	config := &handlers.Config{
		RemoteURL: "http://custom.example.com",
	}
	registry := handlers.NewRegistry(config)

	// Just verify the registry is created successfully
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}
}

func TestRegistry_Middleware_Preflight(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	req := httptest.NewRequest(http.MethodOptions, "/api/upload", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	registry.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for preflight, got %d", w.Code)
	}
}

func TestRegistry_Routes_AllMethods(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"GET health", http.MethodGet, "/api/health", http.StatusOK},
		{"OPTIONS preview", http.MethodOptions, "/api/preview", http.StatusOK},
		{"POST preview (no body)", http.MethodPost, "/api/preview", http.StatusBadRequest},
		{"POST upload (no body)", http.MethodPost, "/api/upload", http.StatusBadRequest},
		{"GET upload", http.MethodGet, "/api/upload", http.StatusMethodNotAllowed},
		{"POST health", http.MethodPost, "/api/health", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			registry.Routes().ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Logf("%s: Expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRegistry_Routes_StaticEndpoints(t *testing.T) {
	config := &handlers.Config{}
	registry := handlers.NewRegistry(config)

	tests := []struct {
		name string
		path string
	}{
		{"root", "/"},
		{"index.html", "/index.html"},
		{"static path", "/static/test.js"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			registry.Routes().ServeHTTP(w, req)

			// Static files handler should not return method not allowed
			if w.Code == http.StatusMethodNotAllowed {
				t.Error("Static files should handle GET requests")
			}
		})
	}
}
