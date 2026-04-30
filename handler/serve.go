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
	if filename == "" {
		http.Error(w, `{"error":"missing filename"}`, http.StatusBadRequest)
		return
	}

	// Reject any path separators and traversal attempts
	if strings.ContainsAny(filename, `/\\`) || strings.Contains(filename, "..") {
		http.Error(w, `{"error":"invalid filename"}`, http.StatusBadRequest)
		return
	}

	// Clean and verify the resolved path is within uploadDir
	cleanName := filepath.Clean(filename)
	if cleanName == "." || cleanName == ".." {
		http.Error(w, `{"error":"invalid filename"}`, http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.uploadDir, cleanName)

	// Final safety check: resolved path must be under uploadDir
	absUploadDir, _ := filepath.Abs(h.uploadDir)
	absFilePath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFilePath, absUploadDir+string(os.PathSeparator)) {
		http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
		return
	}

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
