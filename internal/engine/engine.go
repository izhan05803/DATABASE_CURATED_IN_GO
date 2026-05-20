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
	mu       sync.RWMutex
	data     map[string]types.Record
	index    types.Index
	storage  types.Storage
	file     *storagepkg.FileManager
	pages    *storagepkg.PageManager
	buffer   *storagepkg.BufferPool
	keyPage  map[string]uint32
	filePath string // Path to database file for stats
	// Metrics tracking for observability
	metricsMu   sync.RWMutex
	opsGet      int64 // Total GET operations
	opsSet      int64 // Total SET operations
	opsDelete   int64 // Total DELETE operations
	cacheHits   int64 // Buffer pool cache hits
	cacheMisses int64 // Buffer pool cache misses
	startTime   time.Time
}

// New creates a new database engine
func New() *Engine {
	return &Engine{
		data:      make(map[string]types.Record),
		index:     indexpkg.NewBTree(16),
		keyPage:   make(map[string]uint32),
		startTime: time.Now(),
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
		data:      make(map[string]types.Record),
		index:     indexpkg.NewBTree(16),
		file:      fm,
		filePath:  path,
		keyPage:   make(map[string]uint32),
		startTime: time.Now(),
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
		data:      make(map[string]types.Record),
		storage:   storage,
		index:     index,
		keyPage:   make(map[string]uint32),
		startTime: time.Now(),
	}
}

// Get retrieves a value by key
func (e *Engine) Get(key string) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Track operation
	e.metricsMu.Lock()
	e.opsGet++
	e.metricsMu.Unlock()

	if e.buffer != nil && e.index != nil {
		if pageID, found := e.index.Search(key); found {
			page, err := e.buffer.GetPage(pageID)
			if err == nil {
				// Track cache hit
				e.metricsMu.Lock()
				e.cacheHits++
				e.metricsMu.Unlock()
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

	// Track operation
	e.metricsMu.Lock()
	e.opsSet++
	e.metricsMu.Unlock()

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

	// Track operation
	e.metricsMu.Lock()
	e.opsDelete++
	e.metricsMu.Unlock()

	if e.index != nil {
		_ = e.index.Delete(key)
	}
	delete(e.keyPage, key)

	return nil
}

// Keys returns all keys matching a glob pattern.
// Supports * (zero or more chars) and ? (exactly one char) wildcards.
// Scans all keys and returns matches in sorted order.
func (e *Engine) Keys(pattern string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var keys []string
	for k, r := range e.data {
		if !r.Deleted && matchPattern(k, pattern) {
			keys = append(keys, k)
		}
	}

	// Sort for consistent output
	sort.Strings(keys)
	return keys
}

// matchPattern is a local wrapper around repl.MatchPattern for pattern matching
func matchPattern(key, pattern string) bool {
	n, m := len(key), len(pattern)

	// dp[i][j] = does key[0..i-1] match pattern[0..j-1]
	dp := make([][]bool, n+1)
	for i := range dp {
		dp[i] = make([]bool, m+1)
	}

	// Base case: empty key matches empty pattern
	dp[0][0] = true

	// Handle patterns like * or ** that can match empty key
	for j := 1; j <= m; j++ {
		if pattern[j-1] == '*' {
			dp[0][j] = dp[0][j-1]
		}
	}

	// Fill DP table
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if pattern[j-1] == '*' {
				// * can match zero chars (dp[i][j-1]) or one+ chars (dp[i-1][j])
				dp[i][j] = dp[i][j-1] || dp[i-1][j]
			} else if pattern[j-1] == '?' || pattern[j-1] == key[i-1] {
				// ? matches any single char, or literal must match
				dp[i][j] = dp[i-1][j-1]
			}
		}
	}

	return dp[n][m]
}

// Info returns database statistics
// Info returns comprehensive database statistics for monitoring
func (e *Engine) Info() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Count active records
	count := 0
	totalSize := 0
	for k, r := range e.data {
		if !r.Deleted {
			count++
			totalSize += len(k) + len(r.Value)
		}
	}

	// Get metrics
	e.metricsMu.RLock()
	opsGet := e.opsGet
	opsSet := e.opsSet
	opsDelete := e.opsDelete
	cacheHits := e.cacheHits
	cacheMisses := e.cacheMisses
	startTime := e.startTime
	e.metricsMu.RUnlock()

	// Calculate cache hit rate
	totalAccesses := cacheHits + cacheMisses
	hitRate := 0.0
	if totalAccesses > 0 {
		hitRate = float64(cacheHits) / float64(totalAccesses) * 100
	}

	// Get file size if persisted
	fileSize := int64(0)
	if e.filePath != "" {
		if stat, err := os.Stat(e.filePath); err == nil {
			fileSize = stat.Size()
		}
	}

	// Calculate uptime
	uptime := time.Since(startTime)

	return map[string]interface{}{
		// Storage & Records
		"records":         count,
		"memory_usage_kb": totalSize / 1024,
		"persisted":       e.file != nil,
		"file_size_bytes": fileSize,

		// Operations
		"total_gets":       opsGet,
		"total_sets":       opsSet,
		"total_deletes":    opsDelete,
		"total_operations": opsGet + opsSet + opsDelete,

		// Cache Performance
		"cache_hits":         cacheHits,
		"cache_misses":       cacheMisses,
		"cache_hit_rate_pct": fmt.Sprintf("%.1f%%", hitRate),

		// System Info
		"uptime_seconds": int64(uptime.Seconds()),
		"server_time":    time.Now().Format("2006-01-02 15:04:05"),
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
