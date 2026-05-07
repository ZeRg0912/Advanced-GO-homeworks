package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

var (
	ErrInvalidURL       = errors.New("invalid URL")
	ErrShortIDNotFound  = errors.New("short URL not found")
	ErrShortIDIsEmpty   = errors.New("short URL is empty")
	ErrShortIDCollision = errors.New("could not generate unique short URL")
)

type URLShortener struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]string),
	}
}

// Shorten создает короткий идентификатор для URL.
func (us *URLShortener) Shorten(originalURL string) (string, error) {
	if !isValidURL(originalURL) {
		return "", ErrInvalidURL
	}

	for attempt := 0; attempt < 10; attempt++ {
		shortID := generateShortID()

		us.mu.Lock()
		if _, exists := us.urls[shortID]; !exists {
			us.urls[shortID] = originalURL
			us.mu.Unlock()
			return shortID, nil
		}
		us.mu.Unlock()
	}

	return "", fmt.Errorf("%w after several attempts", ErrShortIDCollision)
}

// GetOriginal возвращает оригинальный URL по короткому ID.
func (us *URLShortener) GetOriginal(shortID string) (string, error) {
	if strings.TrimSpace(shortID) == "" {
		return "", ErrShortIDIsEmpty
	}

	us.mu.RLock()
	originalURL, exists := us.urls[shortID]
	us.mu.RUnlock()

	if !exists {
		return "", ErrShortIDNotFound
	}

	return originalURL, nil
}

// generateShortID генерирует случайный короткий идентификатор длиной 8 символов.
func generateShortID() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("generate short ID: %w", err))
	}

	return base64.RawURLEncoding.EncodeToString(buf)
}

// isValidURL проверяет корректность URL. Валидными считаются только HTTP/HTTPS URL с host.
func isValidURL(str string) bool {
	if str == "" || strings.TrimSpace(str) != str {
		return false
	}

	parsedURL, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if parsedURL.Host == "" || strings.ContainsAny(parsedURL.Host, " \t\r\n") {
		return false
	}

	return true
}
