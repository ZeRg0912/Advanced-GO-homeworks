package http

import (
	"log"
	stdhttp "net/http"
	"time"
)

// LoggingMiddleware logs the incoming request method, path, status and duration.
func LoggingMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		start := time.Now()
		wrapped := &statusResponseWriter{ResponseWriter: w, statusCode: stdhttp.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf("method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

// JSONContentTypeMiddleware sets the response Content-Type for all API responses.
func JSONContentTypeMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// RecoverMiddleware converts unexpected panics into JSON HTTP 500 responses.
func RecoverMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		defer func() {
			if value := recover(); value != nil {
				log.Printf("panic recovered: method=%s path=%s value=%v", r.Method, r.URL.Path, value)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(stdhttp.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal server error"}` + "\n"))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type statusResponseWriter struct {
	stdhttp.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
