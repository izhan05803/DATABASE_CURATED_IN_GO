package storage

import (
	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

const (
	PageSize = 4096 // 4KB pages
)

// PageManager handles page allocation and management
type PageManager struct {
	pages  map[uint32]*types.Page
	nextID uint32
}

// NewPageManager creates a new page manager
func NewPageManager() *PageManager {
	return &PageManager{
		pages:  make(map[uint32]*types.Page),
		nextID: 0,
	}
}

// AllocatePage creates a new page and returns its ID
func (pm *PageManager) AllocatePage(pageType types.PageType) uint32 {
	id := pm.nextID
	pm.nextID++
	pm.pages[id] = &types.Page{
		PageID:   id,
		PageType: pageType,
		Records:  make([]types.Record, 0),
		NextPage: 0,
	}
	return id
}

// GetPage retrieves a page by ID
func (pm *PageManager) GetPage(id uint32) (*types.Page, bool) {
	page, ok := pm.pages[id]
	return page, ok
}

// FreePage removes a page from the manager
func (pm *PageManager) FreePage(id uint32) {
	delete(pm.pages, id)
}
