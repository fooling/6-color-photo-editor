# 6-Color E-Ink Photo Editor

A web-based photo editor designed specifically for 6-color E-Ink displays (Black, White, Red, Green, Blue, Yellow), with a pipeline that handles cropping, enhancement, dithering, palette quantization, and direct upload to the target device over HTTP.

Typical use case: you have an ESP32-driven 6-color E-Ink frame (e.g. a Waveshare 7.3" 800×480 panel), you want to send it a photo, and you need the image pre-processed so the limited palette still looks good.

---

## What's in the box

- **Web UI** — drag-and-drop an image, crop it to the display's aspect ratio, tweak brightness/contrast/saturation, preview the dithered result side-by-side with the original, then push it to the display.
- **CLI** — scriptable single-shot conversion (`convert`) plus the same web server (`server`). Supports stdin pipes.
- **Processing pipeline** — crop → resize → enhance → dither (Floyd–Steinberg) → quantize to 6-color palette.
- **Pluggable enhancers** — three built in (Basic, Auto Levels, E-Ink Optimized); register your own at runtime.
- **HTTP uploader** — encodes to 24-bit BMP and POSTs to the device endpoint.

## Requirements

- Go **1.22+** (see `go.mod`)
- A 6-color E-Ink display reachable over HTTP (optional — the editor is usable without one; you can just download the processed image)

---

## Quick start

### Option A: Build locally and run the web UI

```bash
make local                   # builds ./build/eink-6color for your OS/arch
./build/eink-6color server   # starts on http://0.0.0.0:3000
```

Open <http://127.0.0.1:3000> in a browser, drop in a JPEG/PNG, and follow the UI.

### Option B: One-shot CLI conversion

```bash
./build/eink-6color convert photo.jpg -o output.png -W 800 -H 480
```

Output is a PNG palette-quantized to the 6 display colors. Pipe from stdin with `-`:

```bash
cat photo.jpg | ./build/eink-6color convert - -o output.png -W 800 -H 480
```

### Option C: Convert and upload in one step

```bash
./build/eink-6color convert photo.jpg -o /tmp/out.png \
  -W 800 -H 480 \
  --upload --remote http://192.168.4.1/dataUP
```

`192.168.4.1` is the default AP-mode address most ESP32 firmware exposes. Change it to your device's LAN address if it's joined to your Wi-Fi.

---

## Using the web UI

1. **Upload** — drag a JPEG/PNG into the drop zone (or click to browse).
2. **Orientation toggle** — switches target between 800×480 (landscape) and 480×800 (portrait). The crop box updates to match.
3. **Crop** — drag the crop handles; aspect ratio is locked to the chosen orientation.
4. **Adjust** — sliders for brightness (−1 to +1), contrast (0–2), saturation (0–2), plus a dithering on/off toggle and an enhancer selector.
5. **Update Preview** — re-runs the pipeline and shows each processing step.
6. **Push to Screen** — encodes the final frame as BMP and POSTs it to the **Remote Display URL** field (default `http://192.168.4.1/dataUP`). Edit that field to match your device.

## CLI reference

Run `./build/eink-6color --help` for the full list. Highlights:

### `server`

| Flag | Short | Default | Description |
|---|---|---|---|
| `--port` | `-p` | `3000` | Port to listen on |
| `--host` | `-H` | `0.0.0.0` | Interface to bind |
| `--remote-url` |  | `http://127.0.0.1/dataUP` | Default upload target pre-filled in the UI |

### `convert [input-file]`

| Flag | Short | Default | Description |
|---|---|---|---|
| `--width` | `-W` | `0` | Target width (0 = keep aspect) |
| `--height` | `-H` | `0` | Target height (0 = keep aspect) |
| `--brightness` |  | `0.0` | −1.0 .. 1.0 |
| `--contrast` |  | `1.0` | 0.0 .. 2.0+ |
| `--saturation` |  | `1.0` | 0.0 .. 2.0+ |
| `--dither` | `-d` | `true` | Floyd–Steinberg on/off |
| `--upload` | `-u` | `false` | POST result to `--remote` after converting |
| `--remote` | `-r` | `http://127.0.0.1/dataUP` | Upload target |
| `--output` | `-o` | *(stdout)* | Output PNG path |

Pass `-` as the input file to read from stdin.

---

## E-Ink device side (what the uploader sends)

The uploader POSTs the processed frame to the configured URL. The wire format matches the Waveshare 1.4.0 firmware:

- **Method:** `POST`
- **Content-Type:** `application/octet-stream`
- **Body:** 1 network-mode byte (`0x01` = STA) followed by a 24-bit BMP (bottom-up BGR rows, 4-byte row padding)
- **Expected response:** HTTP 200 (any non-200 is treated as failure and the body is logged)

Your firmware needs to accept that request on whatever path you point the editor at (`/dataUP` for 1.4.0 and newer; pre-1.4.0 used `/esp/dataUP` and raw `image/bmp` — those are not supported by this build). The client always sends `0x01` (STA); AP (`0x00`) is not supported.

---

## Choosing an enhancer

The 6-color palette is harsh — raw photos often look muddy after quantization. Start with the enhancer that matches your source:

- **Basic** — you want manual control. Use the sliders.
- **Auto Levels** — dull/hazy photos with compressed histograms. Stretches black/white points automatically.
- **E-Ink Optimized** — general-purpose "make this look good on the display." Boosts saturation toward the six palette colors and sharpens edges to compensate for dithering softness.

If a result looks washed out, try E-Ink Optimized first. If it's blown out, fall back to Basic with contrast ≈ 1.2 and saturation ≈ 1.3.

---

## Architecture

```
6-color-photo-editor/
├── cmd/                   # Cobra CLI entry points
│   ├── root/              # Root command wiring
│   ├── convert/           # `convert` subcommand
│   └── server/            # `server` subcommand
├── internal/
│   ├── core/
│   │   ├── dither.go      # Floyd–Steinberg implementation
│   │   └── palette/       # 6-color palette + nearest-color matcher
│   ├── pipeline/          # Ordered processing stages
│   │   ├── crop.go        # Cropping (bounds-normalized to (0,0))
│   │   ├── resize.go      # Aspect-aware resize
│   │   ├── enhance*.go    # Basic / Auto Levels / E-Ink enhancers
│   │   ├── enhancer.go    # Enhancer interface + registry
│   │   ├── dither.go      # Pipeline stage wrapping core dither
│   │   └── pipeline.go    # Orchestrator (`ProcessWithConfig`)
│   └── web/handlers/      # HTTP handlers + embedded static assets
│       └── static/        # index.html / styles.css / app.js
└── pkg/
    ├── encoder/           # BMP + raw E-Ink encoder
    └── uploader/          # HTTP client that POSTs the frame
```

### Pipeline stages

The canonical order, as run by `pipeline.ProcessWithConfig`:

```
input → crop → resize → enhance → dither → palette-quantize → final frame
```

Each stage can be inspected individually in the web UI's "Processing Pipeline" panel, which is handy when tuning an image.

### Enhancer plugin interface

Every enhancer implements:

```go
type Enhancer interface {
    Name() string
    DisplayName() string
    Description() string
    Apply(img image.Image) (image.Image, error)
}
```

Register a new one at runtime:

```go
pipeline.RegisterEnhancer(myEnhancer)
```

It'll show up automatically in the web UI's Enhancement Mode dropdown (served by `GET /api/enhancers`).

---

## Development

### Run tests

```bash
make test          # verbose + race + coverage, writes coverage.html
make check         # fmt + vet + test
go test ./...      # quick run
```

### Build all supported platforms

```bash
make build         # linux/{amd64,arm64}, darwin/{amd64,arm64}, windows/{amd64,arm64}
make release       # same + sha256 sums alongside each binary
```

Output lands in `build/` (gitignored).

### Other targets

```bash
make linux         # just the Linux binaries
make local         # current platform only
make lint          # go vet + go fmt + golangci-lint if installed
make clean         # remove build/ and coverage files
make help          # full target list
```

---

## HTTP API reference

| Method | Path | Purpose |
|---|---|---|
| GET | `/` | Web UI (embedded HTML/CSS/JS) |
| GET | `/api/health` | Liveness check |
| GET | `/api/enhancers` | JSON list of registered enhancers |
| POST | `/api/preview` | Process an image and return each pipeline stage as base64 |
| POST | `/api/upload` | Process an image and POST the final frame to the configured device |

Both POST endpoints accept multipart form uploads with the image plus pipeline parameters (width/height/brightness/contrast/saturation/dither/enhancer/remoteURL). See `internal/web/handlers/processor.go` for the exact field list.

---

## Troubleshooting

- **Upload returns non-200** — the uploader logs the full request/response (status, headers, body) to stderr. Check the device's HTTP server logs against that. Most often it's the wrong path, a content-type mismatch, or the device isn't at the configured IP.
- **Web UI shows `192.168.4.1` but my device is on my LAN** — that's just the default placeholder for ESP32 AP mode. Edit the "Remote Display URL" field in the UI, or launch the server with `--remote-url http://<your-ip>/<path>` so it pre-fills with the right value.
- **Preview looks posterized / bandy** — make sure "Dithering" is enabled. Without it, the output is flat palette quantization.
- **Colors look wrong on device but fine in preview** — the device's actual color response is narrower than sRGB. Try the "E-Ink Optimized" enhancer, which compensates.

---

## License

GPL-3.0 — see [LICENSE](LICENSE) for the full text.
