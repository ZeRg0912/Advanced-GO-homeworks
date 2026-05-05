package storage

import (
	"errors"

	"tasks-api/internal/models"
)

var (
	// ErrNotFound is returned when a task with the requested id does not exist.
	ErrNotFound = errors.New("task not found")
)

// Storage describes a task repository used by HTTP handlers.
type Storage interface {
	List() []models.Task
	Create(models.Task) (models.Task, error)
	Get(id int) (models.Task, bool)
	Update(id int, task models.Task) (models.Task, error)
	Delete(id int) error
}
