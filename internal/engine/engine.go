package engine

import (
	"sync"
	"time"

	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

// Engine is the core database engine
type Engine struct {
	mu      sync.RWMutex
	data    map[string]types.Record
	index   types.Index
	storage types.Storage
}

// New creates a new database engine
func New() *Engine {
	return &Engine{
		data: make(map[string]types.Record),
	}
}

// NewWithStorage creates a new engine with a storage backend
func NewWithStorage(storage types.Storage, index types.Index) *Engine {
	return &Engine{
		data:    make(map[string]types.Record),
		storage: storage,
		index:   index,
	}
}

// Get retrieves a value by key
func (e *Engine) Get(key string) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	record, ok := e.data[key]
	if !ok || record.Deleted {
		return nil, types.ErrKeyNotFound
	}

	return record.Value, nil
}

// Set stores a key-value pair
func (e *Engine) Set(key string, value []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.data[key] = types.Record{
		Key:       key,
		Value:     value,
		Timestamp: time.Now().UnixNano(),
		Deleted:   false,
	}

	return nil
}

// Delete removes a key
func (e *Engine) Delete(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	record, ok := e.data[key]
	if !ok {
		return types.ErrKeyNotFound
	}

	record.Deleted = true
	record.Timestamp = time.Now().UnixNano()
	e.data[key] = record

	return nil
}

// Keys returns all keys matching a pattern (simple prefix match for MVP)
func (e *Engine) Keys(pattern string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var keys []string
	for k, r := range e.data {
		if !r.Deleted {
			// Simple implementation: return all keys if pattern is "*"
			// or keys that start with pattern (minus trailing *)
			if pattern == "*" {
				keys = append(keys, k)
			} else if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
				prefix := pattern[:len(pattern)-1]
				if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
					keys = append(keys, k)
				}
			} else if k == pattern {
				keys = append(keys, k)
			}
		}
	}

	return keys
}

// Info returns database statistics
func (e *Engine) Info() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	count := 0
	for _, r := range e.data {
		if !r.Deleted {
			count++
		}
	}

	return map[string]interface{}{
		"records": count,
	}
}

// Close shuts down the engine
func (e *Engine) Close() error {
	if e.storage != nil {
		return e.storage.Close()
	}
	return nil
}
