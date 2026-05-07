package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type appHandler struct {
	shortener *URLShortener
}

func newServer(shortener *URLShortener) http.Handler {
	h := &appHandler{shortener: shortener}

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", h.handleShorten)
	mux.HandleFunc("/", h.handleRedirect)

	return mux
}

func (h *appHandler) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	defer r.Body.Close()

	var req shortenRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON"})
		return
	}

	shortID, err := h.shortener.Shorten(req.URL)
	if err != nil {
		if errors.Is(err, ErrInvalidURL) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid URL"})
			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, shortenResponse{
		ShortURL:    shortID,
		OriginalURL: req.URL,
	})
}

func (h *appHandler) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	shortID := strings.TrimPrefix(r.URL.Path, "/")
	if shortID == "" || strings.Contains(shortID, "/") {
		http.NotFound(w, r)
		return
	}

	originalURL, err := h.shortener.GetOriginal(shortID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func main() {
	shortener := NewURLShortener()
	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", newServer(shortener)))
}
