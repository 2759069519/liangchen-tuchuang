package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"imgbed/store"
)

type ListHandler struct {
	store   *store.Store
	baseURL string
}

func NewListHandler(s *store.Store, baseURL string) *ListHandler {
	return &ListHandler{
		store:   s,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

type ImageItem struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Thumb    string `json:"thumb,omitempty"`
	Size     int64  `json:"size"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type ListResponse struct {
	Items  []ImageItem `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Parse pagination params
	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	images, total, err := h.store.List(limit, offset)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	items := make([]ImageItem, 0, len(images))
	for _, img := range images {
		item := ImageItem{
			ID:       img.ID,
			Filename: img.Filename,
			URL:      h.baseURL + "/img/" + img.StoredName,
			Size:     img.Size,
			Width:    img.Width,
			Height:   img.Height,
		}
		if img.ThumbName != "" {
			item.Thumb = h.baseURL + "/img/" + img.ThumbName
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
