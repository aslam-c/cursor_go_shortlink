package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"
	"url-shortener/database"
)

type URL struct {
	ID          int       `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Clicks      int       `json:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type URLService struct {
	db *database.DB
}

func NewURLService(db *database.DB) *URLService {
	return &URLService{db: db}
}

// CreateShortURL creates a new short URL entry
func (s *URLService) CreateShortURL(originalURL string, customCode string, expiresAt *time.Time) (*URL, error) {
	var shortCode string
	var err error

	if customCode != "" {
		// Use custom code if provided
		shortCode = customCode
		// Check if custom code already exists
		if exists, err := s.shortCodeExists(shortCode); err != nil {
			return nil, err
		} else if exists {
			return nil, fmt.Errorf("custom short code '%s' already exists", shortCode)
		}
	} else {
		// Generate random short code
		shortCode, err = s.generateShortCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %v", err)
		}
	}

	query := `
		INSERT INTO urls (original_url, short_code, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, original_url, short_code, created_at, updated_at, clicks, expires_at
	`

	var url URL
	err = s.db.QueryRow(query, originalURL, shortCode, expiresAt).Scan(
		&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, 
		&url.UpdatedAt, &url.Clicks, &url.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create short URL: %v", err)
	}

	log.Printf("Created short URL: %s -> %s", shortCode, originalURL)
	return &url, nil
}

// GetOriginalURL retrieves the original URL by short code and increments click count
func (s *URLService) GetOriginalURL(shortCode string) (*URL, error) {
	// First, get the URL and check if it exists and hasn't expired
	query := `
		SELECT id, original_url, short_code, created_at, updated_at, clicks, expires_at
		FROM urls 
		WHERE short_code = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`

	var url URL
	err := s.db.QueryRow(query, shortCode).Scan(
		&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt,
		&url.UpdatedAt, &url.Clicks, &url.ExpiresAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("short URL not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve URL: %v", err)
	}

	// Increment click count
	updateQuery := `UPDATE urls SET clicks = clicks + 1 WHERE id = $1`
	_, err = s.db.Exec(updateQuery, url.ID)
	if err != nil {
		log.Printf("Warning: failed to increment click count for URL %d: %v", url.ID, err)
	}

	url.Clicks++ // Update the local copy
	return &url, nil
}

// GetURLStats retrieves URL statistics by short code
func (s *URLService) GetURLStats(shortCode string) (*URL, error) {
	query := `
		SELECT id, original_url, short_code, created_at, updated_at, clicks, expires_at
		FROM urls 
		WHERE short_code = $1
	`

	var url URL
	err := s.db.QueryRow(query, shortCode).Scan(
		&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt,
		&url.UpdatedAt, &url.Clicks, &url.ExpiresAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("short URL not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve URL stats: %v", err)
	}

	return &url, nil
}

// ListURLs retrieves all URLs (with pagination)
func (s *URLService) ListURLs(limit, offset int) ([]URL, error) {
	query := `
		SELECT id, original_url, short_code, created_at, updated_at, clicks, expires_at
		FROM urls 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list URLs: %v", err)
	}
	defer rows.Close()

	var urls []URL
	for rows.Next() {
		var url URL
		err := rows.Scan(&url.ID, &url.OriginalURL, &url.ShortCode, 
			&url.CreatedAt, &url.UpdatedAt, &url.Clicks, &url.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan URL: %v", err)
		}
		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over URLs: %v", err)
	}

	return urls, nil
}

// DeleteURL deletes a URL by short code
func (s *URLService) DeleteURL(shortCode string) error {
	query := `DELETE FROM urls WHERE short_code = $1`
	result, err := s.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("short URL not found")
	}

	log.Printf("Deleted short URL: %s", shortCode)
	return nil
}

// generateShortCode generates a random short code
func (s *URLService) generateShortCode() (string, error) {
	const maxRetries = 10
	
	for i := 0; i < maxRetries; i++ {
		// Generate 6 random bytes
		bytes := make([]byte, 6)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}

		// Encode to base64 and clean up
		shortCode := base64.URLEncoding.EncodeToString(bytes)
		shortCode = strings.TrimRight(shortCode, "=") // Remove padding
		if len(shortCode) > 8 {
			shortCode = shortCode[:8]
		}

		// Check if this code already exists
		exists, err := s.shortCodeExists(shortCode)
		if err != nil {
			return "", err
		}

		if !exists {
			return shortCode, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short code after %d retries", maxRetries)
}

// shortCodeExists checks if a short code already exists in the database
func (s *URLService) shortCodeExists(shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`
	var exists bool
	err := s.db.QueryRow(query, shortCode).Scan(&exists)
	return exists, err
}