package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"

	"imgbed/store"
)

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

type UploadHandler struct {
	store      *store.Store
	uploadDir  string
	maxSize    int64
	baseURL    string
	thumbWidth int
}

func NewUploadHandler(s *store.Store, uploadDir string, maxSize int64, baseURL string) *UploadHandler {
	return &UploadHandler{
		store:      s,
		uploadDir:  uploadDir,
		maxSize:    maxSize,
		baseURL:    strings.TrimRight(baseURL, "/"),
		thumbWidth: 300,
	}
}

type UploadResponse struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Thumb string `json:"thumb,omitempty"`
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)

	// Parse multipart form (max 32MB in memory, rest goes to temp file)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, `{"error":"invalid form data or file too large"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"missing 'file' field"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from extension
		ext := strings.ToLower(filepath.Ext(header.Filename))
		contentType = mime.TypeByExtension(ext)
	}
	if !allowedTypes[contentType] {
		http.Error(w, fmt.Sprintf(`{"error":"unsupported type: %s"}`, contentType), http.StatusBadRequest)
		return
	}

	// Read file into buffer for processing
	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, `{"error":"failed to read file"}`, http.StatusInternalServerError)
		return
	}

	// Decode image to get dimensions
	imgCfg, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		http.Error(w, `{"error":"invalid image data"}`, http.StatusBadRequest)
		return
	}

	// Generate short ID and stored filename
	id := uuid.New().String()[:8]
	ext := formatToExt(format)
	storedName := id + ext
	thumbName := id + "_thumb.jpg"

	// Save original file
	filePath := filepath.Join(h.uploadDir, storedName)
	if err := os.WriteFile(filePath, buf, 0644); err != nil {
		http.Error(w, `{"error":"failed to save file"}`, http.StatusInternalServerError)
		return
	}

	// Generate thumbnail
	thumbPath := filepath.Join(h.uploadDir, thumbName)
	img, err := imaging.Decode(bytes.NewReader(buf))
	if err == nil {
		thumb := imaging.Resize(img, h.thumbWidth, 0, imaging.Lanczos)
		if err := imaging.Save(thumb, thumbPath, imaging.JPEGQuality(80)); err != nil {
			// Thumbnail failure is non-fatal
			thumbName = ""
		}
	} else {
		thumbName = ""
	}

	// Save metadata
	now := time.Now().UTC()
	imageRecord := &store.Image{
		ID:          id,
		Filename:    header.Filename,
		StoredName:  storedName,
		ContentType: contentType,
		Size:        int64(len(buf)),
		Width:       imgCfg.Width,
		Height:      imgCfg.Height,
		ThumbName:   thumbName,
		CreatedAt:   now,
	}

	if err := h.store.Save(imageRecord); err != nil {
		// Cleanup files on DB error
		os.Remove(filePath)
		if thumbName != "" {
			os.Remove(thumbPath)
		}
		http.Error(w, `{"error":"failed to save metadata"}`, http.StatusInternalServerError)
		return
	}

	// Return response
	resp := UploadResponse{
		ID:  id,
		URL: h.baseURL + "/img/" + storedName,
	}
	if thumbName != "" {
		resp.Thumb = h.baseURL + "/img/" + thumbName
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func formatToExt(format string) string {
	switch format {
	case "jpeg":
		return ".jpg"
	case "png":
		return ".png"
	case "gif":
		return ".gif"
	default:
		return ".jpg"
	}
}
