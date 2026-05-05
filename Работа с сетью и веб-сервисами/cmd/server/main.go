package main

import (
	"log"
	"net/http"

	"tasks-api/internal/handlers"
	apphttp "tasks-api/internal/http"
	"tasks-api/internal/storage"
	"tasks-api/internal/storage/memory"
)

func main() {
	var store storage.Storage = memory.New()
	h := handlers.New(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", h.TasksCollection) // GET, POST
	mux.HandleFunc("/tasks/", h.TaskItem)       // GET, PUT, DELETE
	mux.HandleFunc("/health", h.Health)         // GET
	mux.HandleFunc("/", h.NotFound)             // JSON 404 for unknown routes

	handler := apphttp.RecoverMiddleware(
		apphttp.LoggingMiddleware(
			apphttp.JSONContentTypeMiddleware(mux),
		),
	)

	log.Println("server listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
