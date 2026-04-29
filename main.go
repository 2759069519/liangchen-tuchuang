package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"imgbed/handler"
	"imgbed/middleware"
	"imgbed/store"
)

func main() {
	cfg := LoadConfig()

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("failed to create upload dir: %v", err)
	}

	// Init database
	db, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	// Create handlers
	uploadH := handler.NewUploadHandler(db, cfg.UploadDir, cfg.MaxFileSize, cfg.BaseURL)
	serveH := handler.NewServeHandler(cfg.UploadDir)
	deleteH := handler.NewDeleteHandler(db, cfg.UploadDir)
	listH := handler.NewListHandler(db, cfg.BaseURL)

	// Setup routes
	mux := http.NewServeMux()

	// Public: serve images
	mux.Handle("GET /img/", serveH)

	// Public: web UI
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./static/index.html")
	})

	// API routes (auth required)
	mux.Handle("POST /api/upload", uploadH)
	mux.Handle("GET /api/images", listH)
	mux.Handle("DELETE /api/images/", deleteH)

	// Apply middleware: CORS -> Auth -> Handler
	handler := middleware.CORS(
		middleware.Auth(cfg.AuthToken)(mux),
	)

	addr := ":" + cfg.Port
	fmt.Printf("🖼️  imgbed started on http://localhost%s\n", addr)
	fmt.Printf("   Upload dir: %s\n", cfg.UploadDir)
	fmt.Printf("   Max file size: %d MB\n", cfg.MaxFileSize/(1024*1024))
	fmt.Printf("   Auth token: %s\n", maskToken(cfg.AuthToken))

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func maskToken(token string) string {
	if len(token) <= 4 {
		return "****"
	}
	return token[:2] + "***" + token[len(token)-2:]
}
