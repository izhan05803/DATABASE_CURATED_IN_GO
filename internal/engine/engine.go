package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	indexpkg "github.com/izhan05803/gofromscratchdb/internal/index"
	storagepkg "github.com/izhan05803/gofromscratchdb/internal/storage"
	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

// Engine is the core database engine
type Engine struct {
	mu      sync.RWMutex
	data    map[string]types.Record
	index   types.Index
	storage types.Storage
	file    *storagepkg.FileManager
	pages   *storagepkg.PageManager
	buffer  *storagepkg.BufferPool
	keyPage map[string]uint32
}

// New creates a new database engine
func New() *Engine {
	return &Engine{
		data:    make(map[string]types.Record),
		index:   indexpkg.NewBTree(16),
		keyPage: make(map[string]uint32),
	}
}

// NewPersistent creates an engine backed by a database file and loads existing data.
func NewPersistent(path string) (*Engine, error) {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create data directory: %w", err)
		}
	}

	fm, err := storagepkg.NewFileManager(path)
	if err != nil {
		return nil, fmt.Errorf("open database file: %w", err)
	}

	e := &Engine{
		data:    make(map[string]types.Record),
		index:   indexpkg.NewBTree(16),
		file:    fm,
		keyPage: make(map[string]uint32),
	}
	e.pages = storagepkg.NewPageManager(fm)
	e.buffer = storagepkg.NewBufferPool(128, e.pages)

	if err := e.Load(); err != nil {
		fm.Close()
		return nil, fmt.Errorf("load database: %w", err)
	}

	return e, nil
}

// NewWithStorage creates a new engine with a storage backend
func NewWithStorage(storage types.Storage, index types.Index) *Engine {
	return &Engine{
		data:    make(map[string]types.Record),
		storage: storage,
		index:   index,
		keyPage: make(map[string]uint32),
	}
}

// Get retrieves a value by key
func (e *Engine) Get(key string) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.buffer != nil && e.index != nil {
		if pageID, found := e.index.Search(key); found {
			page, err := e.buffer.GetPage(pageID)
			if err == nil {
				for _, rec := range page.Records {
					if rec.Key == key && !rec.Deleted {
						return rec.Value, nil
					}
				}
			}
		}
	}

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

	if e.index != nil {
		if pageID, ok := e.keyPage[key]; ok {
			_ = e.index.Insert(key, pageID)
		}
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

	if e.index != nil {
		_ = e.index.Delete(key)
	}
	delete(e.keyPage, key)

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
		"records":   count,
		"persisted": e.file != nil,
	}
}

// Save serializes all records to disk.
func (e *Engine) Save() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.file == nil || e.pages == nil || e.buffer == nil {
		return nil
	}

	if err := e.pages.Reset(); err != nil {
		return fmt.Errorf("reset pages: %w", err)
	}

	e.index = indexpkg.NewBTree(16)
	e.keyPage = make(map[string]uint32)

	keys := make([]string, 0, len(e.data))
	for k, r := range e.data {
		if !r.Deleted {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	if len(keys) == 0 {
		if err := e.file.SetRootPage(0); err != nil {
			return fmt.Errorf("set root page: %w", err)
		}
		return nil
	}

	var rootPageID uint32
	var prevPageID uint32
	first := true

	for _, key := range keys {
		rec := e.data[key]

		pageID, err := e.pages.AllocatePage(types.PageTypeLeaf)
		if err != nil {
			return fmt.Errorf("allocate page for key %q: %w", key, err)
		}

		page := &types.Page{
			PageID:   pageID,
			PageType: types.PageTypeLeaf,
			Records:  []types.Record{rec},
			NextPage: 0,
		}

		if err := e.buffer.PutPage(page); err != nil {
			return fmt.Errorf("write page for key %q: %w", key, err)
		}

		if e.index != nil {
			if err := e.index.Insert(key, pageID); err != nil {
				return fmt.Errorf("index insert for key %q: %w", key, err)
			}
			e.keyPage[key] = pageID
		}

		if first {
			rootPageID = pageID
			first = false
		} else {
			prev, err := e.buffer.GetPage(prevPageID)
			if err != nil {
				return fmt.Errorf("load previous page %d: %w", prevPageID, err)
			}
			prev.NextPage = pageID
			if err := e.buffer.PutPage(prev); err != nil {
				return fmt.Errorf("link page %d to %d: %w", prevPageID, pageID, err)
			}
		}

		prevPageID = pageID
	}

	if err := e.file.SetRootPage(rootPageID); err != nil {
		return fmt.Errorf("set root page: %w", err)
	}

	if err := e.buffer.Flush(); err != nil {
		return fmt.Errorf("flush buffer pool: %w", err)
	}

	if err := e.file.Sync(); err != nil {
		return fmt.Errorf("sync database file: %w", err)
	}

	return nil
}

// Load deserializes records from disk into memory.
func (e *Engine) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.file == nil || e.pages == nil || e.buffer == nil {
		return nil
	}

	e.data = make(map[string]types.Record)
	e.index = indexpkg.NewBTree(16)
	e.keyPage = make(map[string]uint32)

	if e.file.TotalPages() == 0 {
		return nil
	}

	currentPageID := e.file.RootPage()
	visited := make(map[uint32]struct{})

	for {
		if _, seen := visited[currentPageID]; seen {
			return fmt.Errorf("page chain loop detected at page %d", currentPageID)
		}
		visited[currentPageID] = struct{}{}

		page, err := e.buffer.GetPage(currentPageID)
		if err != nil {
			return fmt.Errorf("read page %d: %w", currentPageID, err)
		}

		for _, rec := range page.Records {
			e.data[rec.Key] = rec
			if !rec.Deleted && e.index != nil {
				if err := e.index.Insert(rec.Key, currentPageID); err != nil {
					return fmt.Errorf("index insert for key %q: %w", rec.Key, err)
				}
				e.keyPage[rec.Key] = currentPageID
			}
		}

		if page.NextPage == 0 {
			break
		}

		currentPageID = page.NextPage
	}

	return nil
}

// Close shuts down the engine
func (e *Engine) Close() error {
	if e.file != nil {
		if e.buffer != nil {
			if err := e.buffer.Flush(); err != nil {
				return err
			}
		}

		if err := e.Save(); err != nil {
			e.file.Close()
			return err
		}
		return e.file.Close()
	}

	if e.storage != nil {
		return e.storage.Close()
	}
	return nil
}
