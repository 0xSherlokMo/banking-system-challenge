// Package memorydb provides an in-memory database implementation that can be used to store and retrieve records.
// It provides methods to lock and unlock records, set and get records, and retrieve all keys and records.
// The MemoryDB type is generic and can store any type that implements the IdentifiedRecord interface.
// The package also defines several errors that can be returned by the methods.
package memorydb

import (
	"errors"
	"sync"
)

type Key = string

// IdentifiedRecord is an interface that must be implemented by any type that is stored in the MemoryDB.
type IdentifiedRecord interface {
	GetID() string
}

var (
	// ErrRowLocked is returned when a row is locked.
	ErrRowLocked = errors.New("row_is_locked")

	// ErrRecordExists is returned when a record already exists.
	ErrRecordExists = errors.New("record_exists")

	// ErrRecordNotFound is returned when a record does not exist.
	ErrRecordNotFound = errors.New("record_not_found")

	// ErrUnlockedBefore is returned when a record is unlocked before trying to unlock it.
	ErrUnlockedBefore = errors.New("unlocked_before")
)

type header struct {
	latch sync.Mutex
}

type Opts struct {
	/*
		Shouldn't be enabled unless locked the key manually, or you want high available endpoint.
	*/
	Safe bool
}

const (
	// Safe, but not highly available.
	ConcurrentSafe = true

	// Highly Available, but run at your own risk.
	ConcurrentNotSafe = false
)

type MemoryDB[T IdentifiedRecord] struct {
	records map[Key]T
	header  map[Key]*header
	setnxmu sync.Mutex
}

func Default[T IdentifiedRecord]() *MemoryDB[T] {
	return &MemoryDB[T]{
		records: make(map[Key]T),
		header:  make(map[Key]*header),
	}
}

// Lock acquires a lock on the given key in the memory database.
// If the key does not exist, it returns an error.
func (m *MemoryDB[T]) Lock(key Key) error {
	pageHeader, exists := m.header[key]
	if !exists {
		return ErrRecordNotFound
	}

	ok := pageHeader.latch.TryLock()
	if !ok {
		return ErrRowLocked
	}
	return nil
}

// Unlock releases a lock on the given key in the memory database.
func (m *MemoryDB[T]) Unlock(key Key) error {
	pageHeader, exists := m.header[key]
	if !exists {
		return ErrRecordNotFound
	}

	ok := pageHeader.latch.TryLock()
	defer pageHeader.latch.Unlock()
	if ok {
		return ErrUnlockedBefore
	}
	return nil
}

// Setnx sets the given key to the given record in the memory database if not exists. otherwise returns an error.
func (m *MemoryDB[T]) Setnx(key Key, record T) error {
	m.setnxmu.Lock()
	defer m.setnxmu.Unlock()
	_, exists := m.header[key]
	if exists {
		return ErrRecordExists
	}

	m.header[key] = new(header)
	m.records[key] = record
	return nil
}

// Set sets the given key to the given record in the memory database.
func (m *MemoryDB[T]) Set(key Key, record T, opts Opts) {
	if !opts.Safe {
		m.records[key] = record
		return
	}

	pageHeader, exists := m.header[key]
	if !exists {
		m.records[key] = record
		m.header[key] = new(header)
		return
	}

	pageHeader.latch.Lock()
	defer pageHeader.latch.Unlock()
	m.records[key] = record
}

// GetM returns the records for the given keys in the memory database.
func (m *MemoryDB[T]) GetM(terms []Key, opts Opts) []T {
	var records []T
	for _, term := range terms {
		record, err := m.Get(term, opts)
		if err != nil {
			continue
		}

		records = append(records, record)
	}

	return records
}

// Get returns the record for the given key in the memory database.
func (m *MemoryDB[T]) Get(key Key, opts Opts) (T, error) {
	if !opts.Safe {
		document, exists := m.records[key]
		if !exists {
			return document, ErrRecordNotFound
		}
		return document, nil
	}

	var record T
	header, exists := m.header[key]
	if !exists {
		return record, ErrRecordNotFound
	}

	header.latch.Lock()
	defer header.latch.Unlock()
	document := m.records[key]
	return document, nil
}

// Keys returns all the keys in the MemoryDB.
func (m *MemoryDB[T]) Keys() []Key {
	var keys []Key
	for k := range m.records {
		keys = append(keys, k)
	}
	return keys
}

// Length returns the number of items in the MemoryDB.
func (m *MemoryDB[T]) Length() int {
	return len(m.records)
}
