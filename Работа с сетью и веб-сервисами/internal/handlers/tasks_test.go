package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tasks-api/internal/handlers"
	apphttp "tasks-api/internal/http"
	"tasks-api/internal/models"
	"tasks-api/internal/storage"
	"tasks-api/internal/storage/memory"
)

type apiError struct {
	Error string `json:"error"`
}

func newTestHandler() http.Handler {
	var store storage.Storage = memory.New()
	h := handlers.New(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", h.TasksCollection)
	mux.HandleFunc("/tasks/", h.TaskItem)
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/", h.NotFound)

	return apphttp.RecoverMiddleware(
		apphttp.LoggingMiddleware(
			apphttp.JSONContentTypeMiddleware(mux),
		),
	)
}

func performRequest(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func requireJSONContentType(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}
}

func decodeResponse[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	var value T
	if err := json.NewDecoder(rec.Body).Decode(&value); err != nil {
		t.Fatalf("decode response: %v; body=%q", err, rec.Body.String())
	}

	return value
}

func TestTaskLifecycle(t *testing.T) {
	handler := newTestHandler()

	listBefore := performRequest(handler, http.MethodGet, "/tasks", "")
	if listBefore.Code != http.StatusOK {
		t.Fatalf("GET /tasks status = %d, want %d", listBefore.Code, http.StatusOK)
	}
	requireJSONContentType(t, listBefore)
	if got := strings.TrimSpace(listBefore.Body.String()); got != "[]" {
		t.Fatalf("GET /tasks body = %q, want []", got)
	}

	createdResp := performRequest(handler, http.MethodPost, "/tasks", `{"title":"Купить продукты","done":false}`)
	if createdResp.Code != http.StatusCreated {
		t.Fatalf("POST /tasks status = %d, want %d; body=%q", createdResp.Code, http.StatusCreated, createdResp.Body.String())
	}
	requireJSONContentType(t, createdResp)
	created := decodeResponse[models.Task](t, createdResp)
	if created.ID != 1 || created.Title != "Купить продукты" || created.Done || created.CreatedAt == "" {
		t.Fatalf("created task = %+v", created)
	}

	gotResp := performRequest(handler, http.MethodGet, "/tasks/1", "")
	if gotResp.Code != http.StatusOK {
		t.Fatalf("GET /tasks/1 status = %d, want %d", gotResp.Code, http.StatusOK)
	}
	requireJSONContentType(t, gotResp)
	got := decodeResponse[models.Task](t, gotResp)
	if got != created {
		t.Fatalf("GET /tasks/1 = %+v, want %+v", got, created)
	}

	updatedResp := performRequest(handler, http.MethodPut, "/tasks/1", `{"title":"Купить продукты и воду","done":true}`)
	if updatedResp.Code != http.StatusOK {
		t.Fatalf("PUT /tasks/1 status = %d, want %d; body=%q", updatedResp.Code, http.StatusOK, updatedResp.Body.String())
	}
	updated := decodeResponse[models.Task](t, updatedResp)
	if updated.ID != 1 || updated.Title != "Купить продукты и воду" || !updated.Done || updated.CreatedAt != created.CreatedAt {
		t.Fatalf("updated task = %+v", updated)
	}

	deletedResp := performRequest(handler, http.MethodDelete, "/tasks/1", "")
	if deletedResp.Code != http.StatusNoContent {
		t.Fatalf("DELETE /tasks/1 status = %d, want %d", deletedResp.Code, http.StatusNoContent)
	}
	requireJSONContentType(t, deletedResp)
	if deletedResp.Body.Len() != 0 {
		t.Fatalf("DELETE /tasks/1 body = %q, want empty", deletedResp.Body.String())
	}

	missingResp := performRequest(handler, http.MethodGet, "/tasks/1", "")
	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("GET deleted /tasks/1 status = %d, want %d", missingResp.Code, http.StatusNotFound)
	}
	missingErr := decodeResponse[apiError](t, missingResp)
	if missingErr.Error != "task not found" {
		t.Fatalf("GET deleted /tasks/1 error = %q", missingErr.Error)
	}
}

func TestValidationAndRoutingErrors(t *testing.T) {
	handler := newTestHandler()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "empty title",
			method:     http.MethodPost,
			path:       "/tasks",
			body:       `{"title":"","done":false}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "title is required",
		},
		{
			name:       "invalid id",
			method:     http.MethodGet,
			path:       "/tasks/abc",
			wantStatus: http.StatusBadRequest,
			wantError:  "task id must be a positive integer",
		},
		{
			name:       "method not allowed",
			method:     http.MethodPatch,
			path:       "/tasks",
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  "method not allowed",
		},
		{
			name:       "unknown route",
			method:     http.MethodGet,
			path:       "/unknown",
			wantStatus: http.StatusNotFound,
			wantError:  "resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performRequest(handler, tt.method, tt.path, tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%q", rec.Code, tt.wantStatus, rec.Body.String())
			}
			requireJSONContentType(t, rec)

			apiErr := decodeResponse[apiError](t, rec)
			if apiErr.Error != tt.wantError {
				t.Fatalf("error = %q, want %q", apiErr.Error, tt.wantError)
			}
		})
	}
}
