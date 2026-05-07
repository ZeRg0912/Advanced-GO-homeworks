package main

import (
	"errors"
	"regexp"
	"sync"
	"testing"
)

func TestURLShortener_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "валидный HTTP URL", url: "http://example.com", wantErr: false},
		{name: "валидный HTTPS URL с query", url: "https://google.com/search?q=test", wantErr: false},
		{name: "валидный HTTPS URL с портом и path", url: "https://example.com:8443/long/path", wantErr: false},
		{name: "невалидный URL без схемы", url: "not-a-url", wantErr: true},
		{name: "пустая строка", url: "", wantErr: true},
		{name: "неподдерживаемая схема", url: "ftp://example.com/file.txt", wantErr: true},
		{name: "нет host", url: "https:///path", wantErr: true},
		{name: "пробелы по краям", url: " https://example.com ", wantErr: true},
	}

	shortener := NewURLShortener()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID, err := shortener.Shorten(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidURL) {
					t.Fatalf("ожидали ErrInvalidURL, получили %v", err)
				}
				return
			}

			if len(shortID) < 6 || len(shortID) > 8 {
				t.Fatalf("короткий ID должен быть длиной 6-8 символов, получили %q длиной %d", shortID, len(shortID))
			}

			gotURL, err := shortener.GetOriginal(shortID)
			if err != nil {
				t.Fatalf("GetOriginal вернул ошибку: %v", err)
			}
			if gotURL != tt.url {
				t.Fatalf("оригинальный URL = %q, ожидали %q", gotURL, tt.url)
			}
		})
	}
}

func TestURLShortener_GetOriginal(t *testing.T) {
	shortener := NewURLShortener()
	originalURL := "https://example.com/long/path"
	shortID, err := shortener.Shorten(originalURL)
	if err != nil {
		t.Fatalf("Shorten вернул ошибку: %v", err)
	}

	tests := []struct {
		name      string
		shortID   string
		wantURL   string
		wantError error
	}{
		{name: "существующий ID", shortID: shortID, wantURL: originalURL, wantError: nil},
		{name: "несуществующий ID", shortID: "unknown", wantURL: "", wantError: ErrShortIDNotFound},
		{name: "пустой ID", shortID: "", wantURL: "", wantError: ErrShortIDIsEmpty},
		{name: "ID из пробелов", shortID: "   ", wantURL: "", wantError: ErrShortIDIsEmpty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := shortener.GetOriginal(tt.shortID)
			if tt.wantError == nil && err != nil {
				t.Fatalf("не ожидали ошибку, получили %v", err)
			}
			if tt.wantError != nil && !errors.Is(err, tt.wantError) {
				t.Fatalf("ошибка = %v, ожидали %v", err, tt.wantError)
			}
			if gotURL != tt.wantURL {
				t.Fatalf("URL = %q, ожидали %q", gotURL, tt.wantURL)
			}
		})
	}
}

func TestGenerateShortID(t *testing.T) {
	idPattern := regexp.MustCompile(`^[A-Za-z0-9_-]{6,8}$`)
	seen := make(map[string]struct{})

	for i := 0; i < 1000; i++ {
		shortID := generateShortID()
		if !idPattern.MatchString(shortID) {
			t.Fatalf("ID %q не соответствует ожидаемому формату", shortID)
		}
		if _, exists := seen[shortID]; exists {
			t.Fatalf("ID %q сгенерирован повторно", shortID)
		}
		seen[shortID] = struct{}{}
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "HTTP URL", url: "http://example.com", want: true},
		{name: "HTTPS URL", url: "https://example.com/path?q=1", want: true},
		{name: "HTTPS localhost с портом", url: "https://localhost:8080/path", want: true},
		{name: "пустая строка", url: "", want: false},
		{name: "URL без схемы", url: "example.com/path", want: false},
		{name: "FTP URL", url: "ftp://example.com", want: false},
		{name: "HTTP без host", url: "http:///path", want: false},
		{name: "URL с пробелом в host", url: "https://exa mple.com", want: false},
		{name: "URL с пробелом в начале", url: " https://example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidURL(tt.url); got != tt.want {
				t.Fatalf("isValidURL(%q) = %v, ожидали %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestURLShortener_ConcurrentShorten(t *testing.T) {
	const goroutines = 50

	shortener := NewURLShortener()
	ids := make(chan string, goroutines)
	errs := make(chan error, goroutines)

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			shortID, err := shortener.Shorten("https://example.com/path")
			if err != nil {
				errs <- err
				return
			}
			ids <- shortID
		}()
	}

	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("Shorten вернул ошибку: %v", err)
		}
	}

	seen := make(map[string]struct{})
	for shortID := range ids {
		if _, exists := seen[shortID]; exists {
			t.Fatalf("получен неуникальный ID: %s", shortID)
		}
		seen[shortID] = struct{}{}
	}

	if len(seen) != goroutines {
		t.Fatalf("получили %d уникальных ID, ожидали %d", len(seen), goroutines)
	}
}
