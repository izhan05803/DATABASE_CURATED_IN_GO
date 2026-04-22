package storage

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

type cacheEntry struct {
	pageID uint32
	page   *types.Page
	dirty  bool
}

// BufferPool keeps a bounded in-memory cache of pages.
type BufferPool struct {
	mu       sync.Mutex
	capacity int
	pages    map[uint32]*list.Element
	lru      *list.List
	manager  *PageManager
}

// NewBufferPool creates a page cache with LRU eviction.
func NewBufferPool(capacity int, manager *PageManager) *BufferPool {
	if capacity <= 0 {
		capacity = 1
	}

	return &BufferPool{
		capacity: capacity,
		pages:    make(map[uint32]*list.Element),
		lru:      list.New(),
		manager:  manager,
	}
}

// GetPage returns a page from cache or loads it from page manager.
func (bp *BufferPool) GetPage(pageID uint32) (*types.Page, error) {
	bp.mu.Lock()
	if elem, ok := bp.pages[pageID]; ok {
		bp.lru.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		page := entry.page
		bp.mu.Unlock()
		return page, nil
	}
	bp.mu.Unlock()

	page, err := bp.manager.ReadPage(pageID)
	if err != nil {
		return nil, err
	}

	bp.mu.Lock()
	defer bp.mu.Unlock()

	if elem, ok := bp.pages[pageID]; ok {
		bp.lru.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		return entry.page, nil
	}

	if err := bp.evictIfNeededLocked(); err != nil {
		return nil, err
	}

	elem := bp.lru.PushFront(&cacheEntry{pageID: pageID, page: page, dirty: false})
	bp.pages[pageID] = elem

	return page, nil
}

// PutPage writes a page through manager and stores it in cache.
func (bp *BufferPool) PutPage(page *types.Page) error {
	if err := bp.manager.WritePage(page); err != nil {
		return err
	}

	bp.mu.Lock()
	defer bp.mu.Unlock()

	if elem, ok := bp.pages[page.PageID]; ok {
		bp.lru.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		entry.page = page
		entry.dirty = false
		return nil
	}

	if err := bp.evictIfNeededLocked(); err != nil {
		return err
	}

	elem := bp.lru.PushFront(&cacheEntry{pageID: page.PageID, page: page, dirty: false})
	bp.pages[page.PageID] = elem

	return nil
}

// DeletePage removes a page from cache and clears it in storage.
func (bp *BufferPool) DeletePage(pageID uint32) error {
	bp.mu.Lock()
	if elem, ok := bp.pages[pageID]; ok {
		bp.lru.Remove(elem)
		delete(bp.pages, pageID)
	}
	bp.mu.Unlock()

	return bp.manager.FreePage(pageID)
}

// Flush ensures dirty pages are persisted.
func (bp *BufferPool) Flush() error {
	bp.mu.Lock()
	entries := make([]*cacheEntry, 0, len(bp.pages))
	for _, elem := range bp.pages {
		entries = append(entries, elem.Value.(*cacheEntry))
	}
	bp.mu.Unlock()

	for _, entry := range entries {
		if entry.dirty {
			if err := bp.manager.WritePage(entry.page); err != nil {
				return err
			}
		}
	}

	return nil
}

// Len returns number of cached pages.
func (bp *BufferPool) Len() int {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return len(bp.pages)
}

func (bp *BufferPool) evictIfNeededLocked() error {
	if len(bp.pages) < bp.capacity {
		return nil
	}

	tail := bp.lru.Back()
	if tail == nil {
		return fmt.Errorf("buffer pool state invalid")
	}

	entry := tail.Value.(*cacheEntry)
	if entry.dirty {
		if err := bp.manager.WritePage(entry.page); err != nil {
			return err
		}
	}

	bp.lru.Remove(tail)
	delete(bp.pages, entry.pageID)
	return nil
}
