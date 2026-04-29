package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ServeHandler struct {
	uploadDir string
}

func NewServeHandler(uploadDir string) *ServeHandler {
	return &ServeHandler{uploadDir: uploadDir}
}

func (h *ServeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract filename from path: /img/{filename}
	filename := strings.TrimPrefix(r.URL.Path, "/img/")
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, `{"error":"invalid filename"}`, http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.uploadDir, filename)

	// Check file exists
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	// Set cache headers (1 year for immutable content-addressed files)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	// Serve the file
	http.ServeFile(w, r, filePath)
}
