package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"time"

	"github.com/fooling/6-color-editor/internal/pipeline"
	"github.com/fooling/6-color-editor/pkg/uploader"
)

// Processor handles image processing requests
type Processor struct{}

// NewProcessor creates a new processor
func NewProcessor() *Processor {
	return &Processor{}
}

// PreviewRequest represents a preview API request
type PreviewRequest struct {
	Image        string  `json:"image"`        // base64 encoded image
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	CropX        float64 `json:"cropX"`        // 0.0 to 1.0
	CropY        float64 `json:"cropY"`        // 0.0 to 1.0
	CropWidth    float64 `json:"cropWidth"`    // 0.0 to 1.0
	CropHeight   float64 `json:"cropHeight"`   // 0.0 to 1.0
	OutputFormat string  `json:"outputFormat"` // "png" or "bmp"
	Brightness   float64 `json:"brightness"`
	Contrast     float64 `json:"contrast"`
	Saturation   float64 `json:"saturation"`
	Dither       bool    `json:"dither"`
	EnhancerName string  `json:"enhancerName"` // enhancer to use (empty = basic with sliders)
}

// PreviewResponse represents a preview API response
type PreviewResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Final string      `json:"final"`    // base64 encoded
		Steps []StepResult `json:"steps"`
	} `json:"result"`
	Error string `json:"error,omitempty"`
}

// StepResult represents a processing step
type StepResult struct {
	Name  string `json:"name"`
	Image string `json:"image"` // base64 encoded thumbnail
}

// UploadRequest represents an upload API request
type UploadRequest struct {
	Image        string  `json:"image"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	CropX        float64 `json:"cropX"`        // 0.0 to 1.0
	CropY        float64 `json:"cropY"`        // 0.0 to 1.0
	CropWidth    float64 `json:"cropWidth"`    // 0.0 to 1.0
	CropHeight   float64 `json:"cropHeight"`   // 0.0 to 1.0
	OutputFormat string  `json:"outputFormat"` // "png" or "bmp"
	Brightness   float64 `json:"brightness"`
	Contrast     float64 `json:"contrast"`
	Saturation   float64 `json:"saturation"`
	Dither       bool    `json:"dither"`
	EnhancerName string  `json:"enhancerName"` // enhancer to use (empty = basic with sliders)
	RemoteURL    string  `json:"remoteUrl,omitempty"`
}

// UploadResponse represents an upload API response
type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// handlePreview processes an image and returns the result with intermediate steps
func (r *Registry) handlePreview(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var previewReq PreviewRequest
	if err := json.NewDecoder(req.Body).Decode(&previewReq); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondJSON(w, PreviewResponse{
			Success: false,
			Error:   "Invalid request body",
		}, http.StatusBadRequest)
		return
	}

	// Decode base64 image
	img, err := decodeBase64Image(previewReq.Image)
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		respondJSON(w, PreviewResponse{
			Success: false,
			Error:   "Failed to decode image",
		}, http.StatusBadRequest)
		return
	}

	// Determine output format (default to PNG)
	outputFormat := previewReq.OutputFormat
	if outputFormat == "" {
		outputFormat = "png"
	}

	// Process image
	config := &pipeline.ProcessConfig{
		Width:        previewReq.Width,
		Height:       previewReq.Height,
		CropX:        previewReq.CropX,
		CropY:        previewReq.CropY,
		CropWidth:    previewReq.CropWidth,
		CropHeight:   previewReq.CropHeight,
		OutputFormat: outputFormat,
		Brightness:   previewReq.Brightness,
		Contrast:     previewReq.Contrast,
		Saturation:   previewReq.Saturation,
		Dither:       previewReq.Dither,
		EnhancerName: previewReq.EnhancerName,
	}

	result := pipeline.ProcessWithConfig(img, config, true)

	// Encode final result
	finalB64, err := encodeImageToBase64(result.Final, outputFormat)
	if err != nil {
		log.Printf("Error encoding result: %v", err)
		respondJSON(w, PreviewResponse{
			Success: false,
			Error:   "Failed to encode result",
		}, http.StatusInternalServerError)
		return
	}

	// Encode intermediate steps
	steps := make([]StepResult, 0, len(result.Steps))
	for _, step := range result.Steps {
		thumbnailB64, err := encodeImageToBase64(step.Image, "png") // Steps always use PNG for preview
		if err != nil {
			log.Printf("Error encoding step %s: %v", step.Name, err)
			continue
		}
		steps = append(steps, StepResult{
			Name:  step.Name,
			Image: thumbnailB64,
		})
	}

	// Set result data
	respData := map[string]any{
		"success": true,
		"result": map[string]any{
			"final": finalB64,
			"steps": steps,
		},
	}

	respondJSON(w, respData, http.StatusOK)
}

// handleUpload processes an image and uploads it to the E-Ink display
func (r *Registry) handleUpload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var uploadReq UploadRequest
	if err := json.NewDecoder(req.Body).Decode(&uploadReq); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondJSON(w, UploadResponse{
			Success: false,
			Error:   "Invalid request body",
		}, http.StatusBadRequest)
		return
	}

	// Decode base64 image
	img, err := decodeBase64Image(uploadReq.Image)
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		respondJSON(w, UploadResponse{
			Success: false,
			Error:   "Failed to decode image",
		}, http.StatusBadRequest)
		return
	}

	// Process image (output format doesn't matter for E-Ink upload, we encode separately)
	config := &pipeline.ProcessConfig{
		Width:        uploadReq.Width,
		Height:       uploadReq.Height,
		CropX:        uploadReq.CropX,
		CropY:        uploadReq.CropY,
		CropWidth:    uploadReq.CropWidth,
		CropHeight:   uploadReq.CropHeight,
		OutputFormat: uploadReq.OutputFormat,
		Brightness:   uploadReq.Brightness,
		Contrast:     uploadReq.Contrast,
		Saturation:   uploadReq.Saturation,
		Dither:       uploadReq.Dither,
		EnhancerName: uploadReq.EnhancerName,
	}

	result := pipeline.ProcessWithConfig(img, config, false)

	// Get uploader
	upldr := r.uploader
	targetURL := "default"
	if uploadReq.RemoteURL != "" {
		targetURL = uploadReq.RemoteURL
		upldr = uploader.NewUploader(&uploader.Config{
			RemoteURL: uploadReq.RemoteURL,
			Timeout:   30 * time.Second,
		})
	}

	// Log upload info
	log.Printf("Uploading to: %s", targetURL)
	log.Printf("Image size: %dx%d", result.Final.Bounds().Dx(), result.Final.Bounds().Dy())

	// Upload
	if err := upldr.Upload(result.Final); err != nil {
		log.Printf("Upload error: %v", err)
		respondJSON(w, UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Upload failed: %v", err),
		}, http.StatusInternalServerError)
		return
	}

	log.Printf("Upload successful to: %s", targetURL)

	respondJSON(w, UploadResponse{
		Success: true,
		Message: "Image uploaded successfully",
	}, http.StatusOK)
}

// handleHealth returns health status
func (r *Registry) handleHealth(w http.ResponseWriter, req *http.Request) {
	respondJSON(w, map[string]any{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	}, http.StatusOK)
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Marshal JSON and write response
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}
	w.Write(bytes)
}

// decodeBase64Image decodes a base64 encoded image
func decodeBase64Image(b64 string) (image.Image, error) {
	// Remove data URL prefix if present
	if len(b64) > 11 && b64[:11] == "data:image/" {
		if idx := findIndex(b64, ","); idx > 0 {
			b64 = b64[idx+1:]
		}
	}

	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	return decodeImageData(data)
}

// encodeImageToBase64 encodes an image to base64 in the specified format
func encodeImageToBase64(img image.Image, format string) (string, error) {
	data, err := encodeImageData(img, format)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// findIndex finds the index of a byte in a string
func findIndex(s string, b string) int {
	for i := range s {
		if s[i:i+1] == b {
			return i
		}
	}
	return -1
}

// decodeImageData decodes image data (supports PNG, JPEG)
func decodeImageData(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// encodeImageData encodes an image to the specified format (png or bmp)
func encodeImageData(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer

	switch format {
	case "bmp":
		// Use BMP encoder
		bmpEncoder := BMPEncoder{}
		data, err := bmpEncoder.Encode(img)
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		// Default to PNG
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}

// BMPEncoder handles BMP encoding
type BMPEncoder struct{}

// Encode encodes an image to BMP format (24-bit RGB)
func (e BMPEncoder) Encode(img image.Image) ([]byte, error) {
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
		Size             uint32
		Width            int32
		Height           int32
		Planes           uint16
		BitsPerPixel     uint16
		Compression      uint32
		ImageSize        uint32
		XPixelsPerM      int32
		YPixelsPerM      int32
		ColorsUsed       uint32
		ColorsImportant  uint32
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

const (
	bmpHeaderSize = 14
	bmpInfoSize   = 40
)
