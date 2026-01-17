# 6-Color E-Ink Photo Editor

A web-based photo editor designed specifically for 6-color E-Ink displays, featuring advanced image processing and real-time preview capabilities.

## Features

### Interactive Crop Tool
- Aspect ratio preservation (800×480 landscape / 480×800 portrait)
- Pixel-perfect resizing with mouse tracking
- Visual feedback with size indicators

### Multiple Enhancement Algorithms

**Basic Enhancer**
- Brightness adjustment (-1.0 to 1.0)
- Contrast control (0.0 to 2.0+)
- Saturation tuning (0.0 to 2.0+)

**Auto Levels**
- Automatic histogram stretching
- Configurable clipping percentage
- Ideal for low-contrast images

**E-Ink Optimized**
- Color separation enhancement
- Edge sharpening for display clarity
- Palette-aware color boosting

### 6-Color Palette
- Black, White, Red, Green, Blue, Yellow
- Floyd-Steinberg dithering for smooth gradients
- Optimized for E-Ink display characteristics

### Real-time Processing
- Live preview with pipeline visualization
- Step-by-step processing display
- Direct HTTP upload to E-Ink display

## Quick Start

### Build
```bash
go build -o 6-color-editor .
```

### Run Server
```bash
./6-color-editor server --port 8080
```

Then open http://127.0.0.1:8080 in your browser.

### Convert Image (CLI)
```bash
./6-color-editor convert input.jpg -o output.bmp -w 800 -h 480
```

## Architecture

```
6-color-editor/
├── cmd/                    # Command-line interfaces
│   ├── convert/           # Image conversion CLI
│   └── server/            # Web server
├── internal/
│   ├── core/              # Core image processing
│   │   └── palette/       # 6-color palette matching
│   ├── pipeline/          # Processing pipeline
│   │   ├── crop.go        # Crop filters
│   │   ├── enhancer*.go   # Enhancement algorithms
│   │   ├── dither.go      # Floyd-Steinberg dithering
│   │   └── resize.go      # Image resizing
│   └── web/               # Web server & handlers
│       └── handlers/
│           └── static/    # Frontend (HTML/CSS/JS)
└── pkg/                   # Public libraries
    ├── encoder/           # BMP encoding
    └── uploader/          # HTTP upload client
```

## Development

### Run Tests
```bash
go test ./...
```

### Test Coverage
```bash
go test -cover ./internal/pipeline/...
```

Current coverage: 83%

### Build with Make
```bash
make build       # Build binary
make test        # Run tests
make coverage    # Generate coverage report
```

## Technical Highlights

### Crop Box Fix
The crop box now correctly maintains aspect ratio in pixel space rather than normalized coordinates, ensuring accurate cropping regardless of the original image's aspect ratio.

### Enhancer Plugin System
Enhancers implement a common interface:
```go
type Enhancer interface {
    Name() string
    DisplayName() string
    Description() string
    Apply(img image.Image) (image.Image, error)
}
```

New enhancers can be registered at runtime:
```go
pipeline.RegisterEnhancer(myEnhancer)
```

### Image Bounds Fix
All crop filters now create images with bounds starting at (0, 0), ensuring compatibility with standard image libraries and preventing coordinate offset bugs.

## API Endpoints

- `GET /` - Web UI
- `POST /api/preview` - Generate preview with processing steps
- `POST /api/upload` - Process and upload to E-Ink display
- `GET /api/enhancers` - List available enhancement algorithms
- `GET /api/health` - Health check

## License

MIT

## Credits

Developed with assistance from Claude Sonnet 4.5
