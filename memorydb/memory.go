package memorydb

import (
	"errors"
	"sync"
)

type Key = string

type IdentifiedRecord interface {
	GetID() string
}

var (
	ErrRowLocked      = errors.New("row_is_locked")
	ErrRecordExists   = errors.New("record_exists")
	ErrRecordNotFound = errors.New("record_not_found")
)

type document[object IdentifiedRecord] struct {
	mu   sync.Mutex
	data object
}

type MemoryDB[T IdentifiedRecord] struct {
	records map[Key]*document[T]
}

func Default[T IdentifiedRecord]() *MemoryDB[T] {
	return &MemoryDB[T]{
		records: make(map[Key]*document[T]),
	}
}

func (m *MemoryDB[T]) Setnx(key Key, record T) error {
	_, exists := m.records[key]
	if exists {
		return ErrRecordExists
	}

	m.records[key] = &document[T]{
		data: record,
	}
	return nil
}

func (m *MemoryDB[T]) Set(key Key, record T) {
	row, exists := m.records[key]
	if !exists {
		m.records[key] = &document[T]{
			data: record,
		}
		return
	}

	row.mu.Lock()
	defer row.mu.Unlock()
	row.data = record
}

func (m *MemoryDB[T]) GetM(terms []Key) []T {
	var records []T
	for _, term := range terms {
		record, err := m.Get(term)
		if err != nil {
			continue
		}

		records = append(records, record)
	}

	return records
}

func (m *MemoryDB[T]) Get(key Key) (T, error) {
	var record T
	row, exists := m.records[key]
	if !exists {
		return record, ErrRecordNotFound
	}

	row.mu.Lock()
	defer row.mu.Unlock()
	return row.data, nil
}

func (m *MemoryDB[T]) Keys() []Key {
	var keys []Key
	for k := range m.records {
		keys = append(keys, k)
	}
	return keys
}

func (m *MemoryDB[T]) Length() int {
	return len(m.records)
}
