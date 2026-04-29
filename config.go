package main

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	UploadDir   string
	DBPath      string
	AuthToken   string
	MaxFileSize int64  // bytes
	BaseURL     string // e.g. http://localhost:8080
}

func LoadConfig() *Config {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		DBPath:      getEnv("DB_PATH", "./imgbed.db"),
		AuthToken:   getEnv("AUTH_TOKEN", "changeme"),
		MaxFileSize: 10 * 1024 * 1024, // 10MB default
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
	}

	if sizeStr := os.Getenv("MAX_FILE_SIZE"); sizeStr != "" {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			cfg.MaxFileSize = size
		}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
