package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Image struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`     // original filename
	StoredName  string    `json:"stored_name"`   // name on disk
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	ThumbName   string    `json:"thumb_name"`    // thumbnail filename, may be empty
	CreatedAt   time.Time `json:"created_at"`
}

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// WAL mode for better concurrent reads
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set wal: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS images (
			id          TEXT PRIMARY KEY,
			filename    TEXT NOT NULL,
			stored_name TEXT NOT NULL,
			content_type TEXT NOT NULL,
			size        INTEGER NOT NULL,
			width       INTEGER DEFAULT 0,
			height      INTEGER DEFAULT 0,
			thumb_name  TEXT DEFAULT '',
			created_at  DATETIME DEFAULT (datetime('now'))
		);
		CREATE INDEX IF NOT EXISTS idx_images_created ON images(created_at DESC);
	`)
	return err
}

func (s *Store) Save(img *Image) error {
	_, err := s.db.Exec(
		`INSERT INTO images (id, filename, stored_name, content_type, size, width, height, thumb_name, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		img.ID, img.Filename, img.StoredName, img.ContentType,
		img.Size, img.Width, img.Height, img.ThumbName, img.CreatedAt,
	)
	return err
}

func (s *Store) Get(id string) (*Image, error) {
	img := &Image{}
	err := s.db.QueryRow(
		`SELECT id, filename, stored_name, content_type, size, width, height, thumb_name, created_at
		 FROM images WHERE id = ?`, id,
	).Scan(&img.ID, &img.Filename, &img.StoredName, &img.ContentType,
		&img.Size, &img.Width, &img.Height, &img.ThumbName, &img.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return img, err
}

func (s *Store) Delete(id string) (string, string, error) {
	// Return stored_name and thumb_name so caller can delete files
	img, err := s.Get(id)
	if err != nil {
		return "", "", err
	}
	if img == nil {
		return "", "", nil
	}

	_, err = s.db.Exec("DELETE FROM images WHERE id = ?", id)
	if err != nil {
		return "", "", err
	}
	return img.StoredName, img.ThumbName, nil
}

func (s *Store) List(limit, offset int) ([]Image, int, error) {
	// Get total count
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM images").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(
		`SELECT id, filename, stored_name, content_type, size, width, height, thumb_name, created_at
		 FROM images ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		if err := rows.Scan(&img.ID, &img.Filename, &img.StoredName, &img.ContentType,
			&img.Size, &img.Width, &img.Height, &img.ThumbName, &img.CreatedAt); err != nil {
			return nil, 0, err
		}
		images = append(images, img)
	}
	return images, total, rows.Err()
}
