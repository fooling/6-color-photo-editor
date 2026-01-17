package handlers

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/fooling/6-color-editor/internal/pipeline"
	"github.com/fooling/6-color-editor/pkg/uploader"
)

//go:embed static/*
var staticFiles embed.FS

// Config holds handler configuration
type Config struct {
	RemoteURL string
}

// Registry manages HTTP handlers and routing
type Registry struct {
	config     *Config
	processor  *Processor
	uploader   *uploader.Uploader
	staticFS   fs.FS
}

// NewRegistry creates a new handler registry
func NewRegistry(config *Config) *Registry {
	if config == nil {
		config = &Config{}
	}

	// Get static files subdirectory
	static, _ := fs.Sub(staticFiles, "static")

	return &Registry{
		config:    config,
		processor: NewProcessor(),
		uploader: uploader.NewUploader(&uploader.Config{
			RemoteURL: config.RemoteURL,
			Timeout:   30 * time.Second,
		}),
		staticFS: static,
	}
}

// Routes returns the main HTTP handler with all routes configured
func (r *Registry) Routes() http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/preview", r.handlePreview)
	mux.HandleFunc("/api/upload", r.handleUpload)
	mux.HandleFunc("/api/health", r.handleHealth)
	mux.HandleFunc("/api/enhancers", r.handleEnhancers)

	// Static files
	mux.Handle("/", http.FileServer(http.FS(r.staticFS)))

	// Apply common middleware
	return r.middleware(mux)
}

// handleEnhancers returns the list of available enhancers
func (r *Registry) handleEnhancers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	enhancers := pipeline.ListEnhancers()
	respondJSON(w, map[string]any{
		"success":   true,
		"enhancers": enhancers,
	}, http.StatusOK)
}

// middleware applies common middleware to the handler
func (r *Registry) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Add common headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		// Disable caching for development
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// CORS for development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
}
