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
"net/http"
"os"
"path/filepath"
"strings"
"time"

"github.com/disintegration/imaging"
"github.com/google/uuid"
_ "golang.org/x/image/webp"

"imgbed/middleware"
"imgbed/store"
)

var allowedTypes = map[string]bool{
"image/jpeg": true,
"image/png":  true,
"image/gif":  true,
"image/webp": true,
}

const guestMaxPerHour = 20

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
ID        string `json:"id"`
URL       string `json:"url"`
Thumb     string `json:"thumb,omitempty"`
Remaining int    `json:"remaining,omitempty"`
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
return
}

isAuth := middleware.IsAuthValid(r)

if !isAuth {
ip := middleware.GetClientIP(r)
allowed, err := h.store.CheckAndIncrementRateLimit(ip, guestMaxPerHour)
if err != nil {
http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
return
}
if !allowed {
remaining, _ := h.store.GetRemainingUploads(ip, guestMaxPerHour)
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusTooManyRequests)
json.NewEncoder(w).Encode(map[string]interface{}{
"error":     "本小时上传次数已达上限，请稍后再试",
"remaining": remaining,
})
return
}
}

r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)

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

buf, err := io.ReadAll(file)
if err != nil {
http.Error(w, `{"error":"failed to read file"}`, http.StatusInternalServerError)
return
}

detectedType := http.DetectContentType(buf)
if !allowedTypes[detectedType] {
http.Error(w, fmt.Sprintf(`{"error":"unsupported type: %s (detected)"}`, detectedType), http.StatusBadRequest)
return
}

imgCfg, format, err := image.DecodeConfig(bytes.NewReader(buf))
if err != nil {
http.Error(w, `{"error":"invalid image data"}`, http.StatusBadRequest)
return
}

id := uuid.New().String()[:8]
ext := formatToExt(format)
storedName := id + ext
thumbName := id + "_thumb.jpg"

filePath := filepath.Join(h.uploadDir, storedName)
if err := os.WriteFile(filePath, buf, 0644); err != nil {
http.Error(w, `{"error":"failed to save file"}`, http.StatusInternalServerError)
return
}

thumbPath := filepath.Join(h.uploadDir, thumbName)
img, err := imaging.Decode(bytes.NewReader(buf))
if err == nil {
thumb := imaging.Resize(img, h.thumbWidth, 0, imaging.Lanczos)
if err := imaging.Save(thumb, thumbPath, imaging.JPEGQuality(80)); err != nil {
thumbName = ""
}
} else {
thumbName = ""
}

now := time.Now().UTC()
imageRecord := &store.Image{
ID:          id,
Filename:    header.Filename,
StoredName:  storedName,
ContentType: detectedType,
Size:        int64(len(buf)),
Width:       imgCfg.Width,
Height:      imgCfg.Height,
ThumbName:   thumbName,
CreatedAt:   now,
}

if err := h.store.Save(imageRecord); err != nil {
os.Remove(filePath)
if thumbName != "" {
os.Remove(thumbPath)
}
http.Error(w, `{"error":"failed to save metadata"}`, http.StatusInternalServerError)
return
}

resp := UploadResponse{
ID:  id,
URL: h.baseURL + "/img/" + storedName,
}
if thumbName != "" {
resp.Thumb = h.baseURL + "/img/" + thumbName
}

if !isAuth {
ip := middleware.GetClientIP(r)
remaining, _ := h.store.GetRemainingUploads(ip, guestMaxPerHour)
resp.Remaining = remaining
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
case "webp":
return ".webp"
default:
return ".jpg"
}
}
