package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BaseEntity là struct cơ bản cho các entity
type BaseEntity struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BaseRepository là repository cơ bản dùng memory
type BaseRepository[T any] struct {
	data map[string]T
	mu   sync.RWMutex
}

// NewBaseRepository khởi tạo
func NewBaseRepository[T any]() *BaseRepository[T] {
	return &BaseRepository[T]{
		data: make(map[string]T),
	}
}

// Create
func (r *BaseRepository[T]) Create(id string, entity T) (T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id == "" {
		id = uuid.New().String()
	}
	// Giả sử entity có ID, CreatedAt, UpdatedAt
	r.data[id] = entity
	return entity, nil
}

// FindAll
func (r *BaseRepository[T]) FindAll() ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]T, 0, len(r.data))
	for _, v := range r.data {
		result = append(result, v)
	}
	return result, nil
}

// FindByID
func (r *BaseRepository[T]) FindByID(id string) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.data[id]
	if !ok {
		var zero T
		return zero, errors.New("not found")
	}
	return v, nil
}

// Update
func (r *BaseRepository[T]) Update(id string, entity T) (T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[id]; !ok {
		var zero T
		return zero, errors.New("not found")
	}
	r.data[id] = entity
	return entity, nil
}

// Delete
func (r *BaseRepository[T]) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[id]; !ok {
		return errors.New("not found")
	}
	delete(r.data, id)
	return nil
}
