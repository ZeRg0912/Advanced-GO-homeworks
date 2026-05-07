package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleShorten(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
		wantURL    string
	}{
		{
			name:       "успешное сокращение HTTP URL",
			method:     http.MethodPost,
			body:       `{"url":"http://example.com/long/path"}`,
			wantStatus: http.StatusOK,
			wantURL:    "http://example.com/long/path",
		},
		{
			name:       "успешное сокращение HTTPS URL",
			method:     http.MethodPost,
			body:       `{"url":"https://example.com/search?q=test"}`,
			wantStatus: http.StatusOK,
			wantURL:    "https://example.com/search?q=test",
		},
		{
			name:       "некорректный JSON",
			method:     http.MethodPost,
			body:       `{"url":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "невалидный URL",
			method:     http.MethodPost,
			body:       `{"url":"not-a-url"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "пустой URL",
			method:     http.MethodPost,
			body:       `{"url":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "лишнее поле в JSON",
			method:     http.MethodPost,
			body:       `{"url":"https://example.com","extra":"value"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "некорректный метод",
			method:     http.MethodGet,
			body:       ``,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newServer(NewURLShortener())
			req := httptest.NewRequest(tt.method, "/shorten", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, ожидали %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var resp shortenResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("не удалось декодировать ответ: %v", err)
			}
			if resp.OriginalURL != tt.wantURL {
				t.Fatalf("original_url = %q, ожидали %q", resp.OriginalURL, tt.wantURL)
			}
			if len(resp.ShortURL) < 6 || len(resp.ShortURL) > 8 {
				t.Fatalf("short_url должен быть длиной 6-8 символов, получили %q", resp.ShortURL)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, ожидали application/json", got)
			}
		})
	}
}

func TestHandleRedirect(t *testing.T) {
	shortener := NewURLShortener()
	originalURL := "https://example.com/long/path"
	shortID, err := shortener.Shorten(originalURL)
	if err != nil {
		t.Fatalf("Shorten вернул ошибку: %v", err)
	}

	tests := []struct {
		name         string
		method       string
		path         string
		wantStatus   int
		wantLocation string
	}{
		{
			name:         "существующий short_url возвращает redirect",
			method:       http.MethodGet,
			path:         "/" + shortID,
			wantStatus:   http.StatusFound,
			wantLocation: originalURL,
		},
		{
			name:       "несуществующий short_url возвращает 404",
			method:     http.MethodGet,
			path:       "/unknown",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "отсутствующий short_url возвращает 404",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "path с вложенным сегментом возвращает 404",
			method:     http.MethodGet,
			path:       "/abc/def",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "некорректный метод возвращает 405",
			method:     http.MethodPost,
			path:       "/" + shortID,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newServer(shortener)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, ожидали %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantLocation != "" && rec.Header().Get("Location") != tt.wantLocation {
				t.Fatalf("Location = %q, ожидали %q", rec.Header().Get("Location"), tt.wantLocation)
			}
		})
	}
}

func TestShortenAndRedirectIntegration(t *testing.T) {
	server := newServer(NewURLShortener())
	originalURL := "https://example.com/some/long/path"

	shortenReq := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(`{"url":"`+originalURL+`"}`))
	shortenRec := httptest.NewRecorder()
	server.ServeHTTP(shortenRec, shortenReq)

	if shortenRec.Code != http.StatusOK {
		t.Fatalf("POST /shorten status = %d, ожидали %d, body = %s", shortenRec.Code, http.StatusOK, shortenRec.Body.String())
	}

	var resp shortenResponse
	if err := json.NewDecoder(shortenRec.Body).Decode(&resp); err != nil {
		t.Fatalf("не удалось декодировать ответ: %v", err)
	}

	redirectReq := httptest.NewRequest(http.MethodGet, "/"+resp.ShortURL, nil)
	redirectRec := httptest.NewRecorder()
	server.ServeHTTP(redirectRec, redirectReq)

	if redirectRec.Code != http.StatusFound {
		t.Fatalf("GET /{short_url} status = %d, ожидали %d", redirectRec.Code, http.StatusFound)
	}
	if redirectRec.Header().Get("Location") != originalURL {
		t.Fatalf("Location = %q, ожидали %q", redirectRec.Header().Get("Location"), originalURL)
	}
}
