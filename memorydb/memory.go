package memorydb

import (
	"errors"
	"sync"
)

type IdentifiedRecord interface {
	GetID() string
}

var ErrRowLocked = errors.New("row_is_locked")

type document[object IdentifiedRecord] struct {
	mu   sync.Mutex
	data object
}

type MemoryDB[T IdentifiedRecord] struct {
	records map[string]*document[T]
}

func Default[T IdentifiedRecord]() *MemoryDB[T] {
	return &MemoryDB[T]{
		records: make(map[string]*document[T]),
	}
}

func (m *MemoryDB[T]) Set(key string, record T) error {
	row, exists := m.records[key]
	if exists {
		row.mu.Lock()
		defer row.mu.Unlock()
	}
	m.records[key] = &document[T]{
		data: record,
	}
	return nil
}

func (m *MemoryDB[T]) Keys() []string {
	var keys []string
	for k := range m.records {
		keys = append(keys, k)
	}
	return keys
}
