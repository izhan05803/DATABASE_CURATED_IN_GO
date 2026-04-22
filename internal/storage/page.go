package storage

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

const (
	PageSize = 4096 // 4KB pages
)

const (
	pageHeaderSize = 16
)

// SerializePage converts a page into a fixed-size 4KB binary representation.
func SerializePage(page *types.Page) ([]byte, error) {
	buf := make([]byte, PageSize)

	binary.LittleEndian.PutUint32(buf[0:4], page.PageID)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(page.PageType))
	binary.LittleEndian.PutUint32(buf[8:12], page.NextPage)
	binary.LittleEndian.PutUint32(buf[12:16], uint32(len(page.Records)))

	offset := pageHeaderSize
	for _, record := range page.Records {
		key := []byte(record.Key)
		value := record.Value

		recordSize := 4 + 4 + 8 + 1 + len(key) + len(value)
		if offset+recordSize > PageSize {
			return nil, types.ErrPageFull
		}

		binary.LittleEndian.PutUint32(buf[offset:offset+4], uint32(len(key)))
		offset += 4

		binary.LittleEndian.PutUint32(buf[offset:offset+4], uint32(len(value)))
		offset += 4

		binary.LittleEndian.PutUint64(buf[offset:offset+8], uint64(record.Timestamp))
		offset += 8

		if record.Deleted {
			buf[offset] = 1
		} else {
			buf[offset] = 0
		}
		offset++

		copy(buf[offset:offset+len(key)], key)
		offset += len(key)

		copy(buf[offset:offset+len(value)], value)
		offset += len(value)
	}

	return buf, nil
}

// DeserializePage reads a fixed-size 4KB binary page into a Page struct.
func DeserializePage(data []byte) (*types.Page, error) {
	if len(data) != PageSize {
		return nil, fmt.Errorf("invalid page size: %d", len(data))
	}

	page := &types.Page{
		PageID:   binary.LittleEndian.Uint32(data[0:4]),
		PageType: types.PageType(binary.LittleEndian.Uint32(data[4:8])),
		NextPage: binary.LittleEndian.Uint32(data[8:12]),
	}

	recordCount := int(binary.LittleEndian.Uint32(data[12:16]))
	page.Records = make([]types.Record, 0, recordCount)

	offset := pageHeaderSize
	for i := 0; i < recordCount; i++ {
		if offset+17 > PageSize {
			return nil, fmt.Errorf("record metadata out of bounds")
		}

		keyLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4

		valueLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4

		timestamp := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
		offset += 8

		deleted := data[offset] == 1
		offset++

		if keyLen < 0 || valueLen < 0 || offset+keyLen+valueLen > PageSize {
			return nil, fmt.Errorf("record payload out of bounds")
		}

		key := string(data[offset : offset+keyLen])
		offset += keyLen

		value := make([]byte, valueLen)
		copy(value, data[offset:offset+valueLen])
		offset += valueLen

		page.Records = append(page.Records, types.Record{
			Key:       key,
			Value:     value,
			Timestamp: timestamp,
			Deleted:   deleted,
		})
	}

	return page, nil
}

// PageManager handles page allocation and management
type PageManager struct {
	mu       sync.RWMutex
	file     *FileManager
	pages    map[uint32]*types.Page
	nextID   uint32
	freeList []uint32
}

// NewPageManager creates a new page manager
func NewPageManager(file *FileManager) *PageManager {
	nextID := uint32(0)
	if file != nil {
		nextID = file.TotalPages()
	}

	return &PageManager{
		file:   file,
		pages:  make(map[uint32]*types.Page),
		nextID: nextID,
	}
}

// AllocatePage creates a new page and returns its ID
func (pm *PageManager) AllocatePage(pageType types.PageType) (uint32, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var id uint32
	if len(pm.freeList) > 0 {
		id = pm.freeList[len(pm.freeList)-1]
		pm.freeList = pm.freeList[:len(pm.freeList)-1]
	} else {
		id = pm.nextID
		pm.nextID++
		if pm.file != nil {
			if err := pm.file.SetTotalPages(pm.nextID); err != nil {
				return 0, err
			}
		}
	}

	page := &types.Page{
		PageID:   id,
		PageType: pageType,
		Records:  make([]types.Record, 0),
		NextPage: 0,
	}

	pm.pages[id] = page
	if pm.file != nil {
		if err := pm.file.WritePage(page); err != nil {
			return 0, err
		}
	}

	return id, nil
}

// GetPage retrieves a page by ID
func (pm *PageManager) GetPage(id uint32) (*types.Page, bool) {
	page, err := pm.ReadPage(id)
	if err != nil {
		return nil, false
	}
	return page, true
}

// ReadPage reads a page from memory cache or file.
func (pm *PageManager) ReadPage(id uint32) (*types.Page, error) {
	pm.mu.RLock()
	if page, ok := pm.pages[id]; ok {
		pm.mu.RUnlock()
		return page, nil
	}
	pm.mu.RUnlock()

	if pm.file == nil {
		return nil, fmt.Errorf("page %d not found", id)
	}

	page, err := pm.file.ReadPage(id)
	if err != nil {
		return nil, err
	}

	pm.mu.Lock()
	pm.pages[id] = page
	pm.mu.Unlock()

	return page, nil
}

// WritePage writes a page to memory cache and file.
func (pm *PageManager) WritePage(page *types.Page) error {
	pm.mu.Lock()
	pm.pages[page.PageID] = page
	pm.mu.Unlock()

	if pm.file != nil {
		return pm.file.WritePage(page)
	}

	return nil
}

// FreePage removes a page from the manager
func (pm *PageManager) FreePage(id uint32) error {
	pm.mu.Lock()
	delete(pm.pages, id)
	pm.freeList = append(pm.freeList, id)
	pm.mu.Unlock()

	if pm.file != nil {
		return pm.file.ClearPage(id)
	}

	return nil
}

// TotalPages returns the page count tracked by the manager.
func (pm *PageManager) TotalPages() uint32 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.nextID
}

// Reset clears all pages and starts allocation from page zero.
func (pm *PageManager) Reset() error {
	pm.mu.Lock()
	pm.pages = make(map[uint32]*types.Page)
	pm.nextID = 0
	pm.freeList = nil
	pm.mu.Unlock()

	if pm.file != nil {
		return pm.file.ResetPages()
	}

	return nil
}
