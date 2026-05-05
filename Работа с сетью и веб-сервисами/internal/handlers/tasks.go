package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"tasks-api/internal/models"
	"tasks-api/internal/storage"
)

type Handler struct{ Store storage.Storage }

func New(s storage.Storage) *Handler { return &Handler{Store: s} }

type taskRequest struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status string `json:"status"`
}

// TasksCollection handles /tasks requests: GET for list and POST for creation.
func (h *Handler) TasksCollection(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/tasks" {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// TaskItem handles /tasks/{id} requests: GET, PUT and DELETE.
func (h *Handler) TaskItem(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, id)
	case http.MethodPut:
		h.updateTask(w, r, id)
	case http.MethodDelete:
		h.deleteTask(w, r, id)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPut, http.MethodDelete}, ", "))
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// NotFound handles unknown routes and keeps the API error format JSON-only.
func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "resource not found")
}

// Health handles GET /health requests.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/health" {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func (h *Handler) listTasks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.Store.List())
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	input, ok := readTaskRequest(w, r)
	if !ok {
		return
	}

	task, err := h.Store.Create(models.Task{
		Title: input.Title,
		Done:  input.Done,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *Handler) getTask(w http.ResponseWriter, _ *http.Request, id int) {
	task, ok := h.Store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request, id int) {
	input, ok := readTaskRequest(w, r)
	if !ok {
		return
	}

	task, err := h.Store.Update(id, models.Task{
		Title: input.Title,
		Done:  input.Done,
	})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) deleteTask(w http.ResponseWriter, _ *http.Request, id int) {
	if err := h.Store.Delete(id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseTaskID(w http.ResponseWriter, r *http.Request) (int, bool) {
	const prefix = "/tasks/"

	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeError(w, http.StatusNotFound, "resource not found")
		return 0, false
	}

	rawID := strings.TrimPrefix(r.URL.Path, prefix)
	if rawID == "" {
		writeError(w, http.StatusBadRequest, "task id is required")
		return 0, false
	}

	if strings.Contains(rawID, "/") {
		writeError(w, http.StatusNotFound, "resource not found")
		return 0, false
	}

	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "task id must be a positive integer")
		return 0, false
	}

	return id, true
}

func readTaskRequest(w http.ResponseWriter, r *http.Request) (taskRequest, bool) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var input taskRequest
	if err := decoder.Decode(&input); err != nil {
		if errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "request body is required")
			return taskRequest{}, false
		}

		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return taskRequest{}, false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		return taskRequest{}, false
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return taskRequest{}, false
	}

	return input, true
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if statusCode == http.StatusNoContent {
		return
	}

	if err := json.NewEncoder(w).Encode(value); err != nil {
		// The response status was already written. Nothing useful can be returned here.
		return
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, errorResponse{Error: message})
}
