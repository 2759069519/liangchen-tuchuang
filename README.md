# imgbed

A simple, self-hosted image hosting service written in pure Go.

## Features

- 📤 Upload images via API or web UI (drag & drop)
- 🔗 Direct URL access to uploaded images
- 🖼️ Auto-generated thumbnails
- 📋 Image listing with pagination
- 🗑️ Image deletion
- 🔑 Bearer token authentication
- 🌐 CORS support

## Quick Start

```bash
# Build
go build -o imgbed .

# Run (default: http://localhost:8080)
./imgbed

# Or with custom config
AUTH_TOKEN=mysecret PORT=9000 BASE_URL=http://myhost:9000 ./imgbed
```

## Configuration

All config is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `UPLOAD_DIR` | `./uploads` | Where to store images |
| `DB_PATH` | `./imgbed.db` | SQLite database path |
| `AUTH_TOKEN` | `changeme` | API authentication token |
| `MAX_FILE_SIZE` | `10485760` | Max upload size in bytes (10MB) |
| `BASE_URL` | `http://localhost:8080` | Base URL for returned links |

## API

### Upload Image

```bash
curl -X POST http://localhost:8080/api/upload \
  -H "Authorization: Bearer your_token" \
  -F "file=@photo.jpg"
```

Response:
```json
{
  "id": "a1b2c3d4",
  "url": "http://localhost:8080/img/a1b2c3d4.jpg",
  "thumb": "http://localhost:8080/img/a1b2c3d4_thumb.jpg"
}
```

### List Images

```bash
curl http://localhost:8080/api/images?limit=20&offset=0 \
  -H "Authorization: Bearer your_token"
```

### Delete Image

```bash
curl -X DELETE http://localhost:8080/api/images/a1b2c3d4 \
  -H "Authorization: Bearer your_token"
```

### Access Image

Just open the URL in a browser:
```
http://localhost:8080/img/a1b2c3d4.jpg
```

## Supported Formats

- JPEG
- PNG
- GIF
- WebP

## Project Structure

```
imgbed/
├── main.go              # Entry point
├── config.go            # Configuration
├── handler/
│   ├── upload.go        # Upload handler
│   ├── serve.go         # Image serving
│   ├── delete.go        # Delete handler
│   └── list.go          # List handler
├── middleware/
│   ├── auth.go          # Token authentication
│   └── cors.go          # CORS
├── store/
│   └── store.go         # SQLite storage layer
├── static/
│   └── index.html       # Web UI
└── uploads/             # Image storage (auto-created)
```

## Build for Other Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o imgbed-linux .

# Windows
GOOS=windows GOARCH=amd64 go build -o imgbed.exe .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o imgbed-mac .
```
