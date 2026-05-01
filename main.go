package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"imgbed/auth"
	"imgbed/handler"
	"imgbed/middleware"
	"imgbed/store"
)

type Config struct {
	Port          string
	UploadDir     string
	DBPath        string
	AuthToken     string
	AdminPassword string
	MaxFileSize   int64
	BaseURL       string
}

func LoadConfig() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		DBPath:        getEnv("DB_PATH", "./imgbed.db"),
		AuthToken:     getEnv("AUTH_TOKEN", ""),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
		MaxFileSize:   getEnvInt64("MAX_FILE_SIZE", 10*1024*1024),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		var n int64
		for _, c := range v {
			if c >= '0' && c <= '9' {
				n = n*10 + int64(c-'0')
			}
		}
		if n > 0 {
			return n
		}
	}
	return fallback
}

func main() {
	cfg := LoadConfig()

	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("failed to create upload dir: %v", err)
	}

	db, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	auth.InitTokenSecret()

	passwordHash := ""
	if cfg.AdminPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("failed to hash password: %v", err)
		}
		passwordHash = string(hash)
	}

	uploadH := handler.NewUploadHandler(db, cfg.UploadDir, cfg.MaxFileSize, cfg.BaseURL)
	serveH := handler.NewServeHandler(cfg.UploadDir)
	deleteH := handler.NewDeleteHandler(db, cfg.UploadDir)
	listH := handler.NewListHandler(db, cfg.BaseURL)
	loginH := handler.NewLoginHandler(passwordHash)

	mux := http.NewServeMux()
	mux.Handle("GET /img/", serveH)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./static/index.html")
	})
	mux.Handle("POST /api/upload", uploadH)
	mux.Handle("GET /api/images", listH)
	mux.Handle("DELETE /api/images/", deleteH)
	mux.Handle("POST /api/login", loginH)

	h := middleware.Logging(
		middleware.CORS(
			middleware.Auth(cfg.AuthToken)(mux),
		),
	)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	fmt.Printf("imgbed started on http://localhost:%s\n", cfg.Port)
	fmt.Printf("  Upload dir: %s\n", cfg.UploadDir)
	fmt.Printf("  Max file size: %d MB\n", cfg.MaxFileSize/(1024*1024))
	if cfg.AuthToken != "" {
		fmt.Printf("  Auth token: %s\n", cfg.AuthToken[:2]+"***"+cfg.AuthToken[len(cfg.AuthToken)-2:])
	}
	if passwordHash != "" {
		fmt.Printf("  Admin login: enabled (bcrypt)\n")
	}
	fmt.Printf("  Guest upload limit: 20/hour per IP\n")

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received signal %v, shutting down...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	log.Printf("server stopped")
}
