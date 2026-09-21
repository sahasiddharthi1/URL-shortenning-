package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"strings"
	"url-shortener/internal/models"
	"url-shortener/internal/store"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	ErrInvalidURL  = errors.New("invalid URL")
	ErrURLNotFound = errors.New("URL not found")
	ErrCodeExists  = errors.New("short code already exists")
)

func getBaseURL() string {
	if url := os.Getenv("RENDER_EXTERNAL_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

type URLService struct {
	store *store.PostgresStore
}

func NewURLService(store *store.PostgresStore) *URLService {
	return &URLService{store: store}
}

func (s *URLService) ShortenURL(longURL string) (*models.ShortenResponse, error) {
	longURL = strings.TrimSpace(longURL)
	if longURL == "" {
		return nil, ErrInvalidURL
	}

	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "https://" + longURL
	}

	// CHECK 1: Does this URL already exist?
	existing, err := s.store.GetURLByLongURL(longURL)
	if err == nil {
		// URL exists → return existing short code
		return &models.ShortenResponse{
			ShortURL:  fmt.Sprintf("%s/%s", getBaseURL(), existing.ShortCode),
			ShortCode: existing.ShortCode,
		}, nil
	}

	// URL doesn't exist → create new short code
	shortCode := s.generateCode(longURL)

	// CHECK 2: Does this code collide with different URL?
	exists, err := s.store.CodeExists(shortCode)
	if err != nil {
		return nil, err
	}
	if exists {
		shortCode = s.generateCodeWithSalt(longURL)
	}

	_, err = s.store.CreateURL(shortCode, longURL)
	if err != nil {
		return nil, err
	}

	return &models.ShortenResponse{
		ShortURL:  fmt.Sprintf("%s/%s", getBaseURL(), shortCode),
		ShortCode: shortCode,
	}, nil
}

func (s *URLService) RedirectURL(shortCode string) (string, error) {
	url, err := s.store.GetURLByCode(shortCode)
	if err != nil {
		return "", ErrURLNotFound
	}
	return url.LongURL, nil
}

func (s *URLService) GetStats(shortCode string) (*models.StatsResponse, []models.AnalyticsEntry, error) {
	return s.store.GetStats(shortCode)
}

func (s *URLService) LogClick(shortCode, referrer, userAgent, ipAddress string) error {
	return s.store.LogAnalytics(shortCode, referrer, userAgent, ipAddress)
}

func (s *URLService) generateCode(longURL string) string {
	hash := md5.Sum([]byte(longURL))
	return s.base62Encode(hash[:7])
}

func (s *URLService) generateCodeWithSalt(longURL string) string {
	for i := 0; i < 100; i++ {
		salted := fmt.Sprintf("%s%d", longURL, i)
		hash := md5.Sum([]byte(salted))
		code := s.base62Encode(hash[:7])
		exists, _ := s.store.CodeExists(code)
		if !exists {
			return code
		}
	}
	return s.generateCode(longURL)
}

func (s *URLService) base62Encode(data []byte) string {
	var result strings.Builder
	for _, b := range data {
		result.WriteByte(chars[int(b)%62])
	}
	return result.String()[:7]
}
