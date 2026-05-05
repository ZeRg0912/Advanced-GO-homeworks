package memory

import (
	"sort"
	"sync"
	"time"

	"tasks-api/internal/models"
	"tasks-api/internal/storage"
)

// Storage is a concurrency-safe in-memory implementation of storage.Storage.
type Storage struct {
	mu     sync.RWMutex
	nextID int
	tasks  map[int]models.Task
}

// New creates an empty in-memory task storage.
func New() *Storage {
	return &Storage{
		nextID: 1,
		tasks:  make(map[int]models.Task),
	}
}

// List returns all tasks sorted by id.
func (s *Storage) List() []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		result = append(result, task)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

// Create stores a task and generates server-side fields.
func (s *Storage) Create(task models.Task) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.ID = s.nextID
	s.nextID++

	if task.CreatedAt == "" {
		task.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	s.tasks[task.ID] = task
	return task, nil
}

// Get returns a task by id.
func (s *Storage) Get(id int) (models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	return task, ok
}

// Update replaces user-controlled task fields and keeps server-controlled fields.
func (s *Storage) Update(id int, task models.Task) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.tasks[id]
	if !ok {
		return models.Task{}, storage.ErrNotFound
	}

	task.ID = id
	task.CreatedAt = existing.CreatedAt

	s.tasks[id] = task
	return task, nil
}

// Delete removes a task by id.
func (s *Storage) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return storage.ErrNotFound
	}

	delete(s.tasks, id)
	return nil
}
