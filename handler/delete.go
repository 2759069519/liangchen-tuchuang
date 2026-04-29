package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"imgbed/store"
)

type DeleteHandler struct {
	store     *store.Store
	uploadDir string
}

func NewDeleteHandler(s *store.Store, uploadDir string) *DeleteHandler {
	return &DeleteHandler{store: s, uploadDir: uploadDir}
}

type DeleteResponse struct {
	Message string `json:"message"`
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path: /api/images/{id}
	id := strings.TrimPrefix(r.URL.Path, "/api/images/")
	if id == "" {
		http.Error(w, `{"error":"missing image id"}`, http.StatusBadRequest)
		return
	}

	storedName, thumbName, err := h.store.Delete(id)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if storedName == "" {
		http.Error(w, `{"error":"image not found"}`, http.StatusNotFound)
		return
	}

	// Delete files
	os.Remove(filepath.Join(h.uploadDir, storedName))
	if thumbName != "" {
		os.Remove(filepath.Join(h.uploadDir, thumbName))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DeleteResponse{Message: "deleted"})
}
