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
