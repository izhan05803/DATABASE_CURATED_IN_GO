package storage

import (
	"os"
	"testing"

	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

// TestBufferPool_GetPage tests retrieving a page from buffer pool
func TestBufferPool_GetPage(t *testing.T) {
	pm := NewPageManager(nil)
	bp := NewBufferPool(10, pm)

	// Allocate and put a page
	id, err := pm.AllocatePage(types.PageTypeLeaf)
	if err != nil {
		t.Fatalf("AllocatePage() error = %v", err)
	}

	page := &types.Page{
		PageID:   id,
		PageType: types.PageTypeLeaf,
		Records:  []types.Record{{Key: "test", Value: []byte("data")}},
	}

	if err := bp.PutPage(page); err != nil {
		t.Fatalf("PutPage() error = %v", err)
	}

	// Get the page back
	got, err := bp.GetPage(id)
	if err != nil {
		t.Fatalf("GetPage() error = %v", err)
	}

	if got.PageID != id || len(got.Records) != 1 {
		t.Errorf("GetPage() mismatch: PageID=%d (want %d), records=%d (want 1)", got.PageID, id, len(got.Records))
	}
}

// TestBufferPool_GetPageNotFound tests getting a page that doesn't exist
func TestBufferPool_GetPageNotFound(t *testing.T) {
	pm := NewPageManager(nil)
	bp := NewBufferPool(10, pm)

	// Try to get non-existent page
	_, err := bp.GetPage(999)
	if err == nil {
		t.Error("GetPage() for non-existent page should return error")
	}
}

// TestBufferPool_DeletePage tests deleting a page from buffer pool
func TestBufferPool_DeletePage(t *testing.T) {
	pm := NewPageManager(nil)
	bp := NewBufferPool(10, pm)

	// Allocate and put a page
	id, err := pm.AllocatePage(types.PageTypeLeaf)
	if err != nil {
		t.Fatalf("AllocatePage() error = %v", err)
	}

	page := &types.Page{PageID: id, PageType: types.PageTypeLeaf}
	if err := bp.PutPage(page); err != nil {
		t.Fatalf("PutPage() error = %v", err)
	}

	// Verify page exists
	if _, err := bp.GetPage(id); err != nil {
		t.Fatalf("GetPage() before delete error = %v", err)
	}

	// Delete the page
	if err := bp.DeletePage(id); err != nil {
		t.Fatalf("DeletePage() error = %v", err)
	}

	// Verify page is gone
	_, err = bp.GetPage(id)
	if err == nil {
		t.Error("GetPage() after delete should return error")
	}
}

// TestBufferPool_Flush tests flushing all pages to disk
func TestBufferPool_Flush(t *testing.T) {
	tmpFile := "test_flush.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	pm := NewPageManager(fm)
	bp := NewBufferPool(2, pm)

	// Add pages to buffer pool
	pages := make([]uint32, 0)
	for i := 0; i < 3; i++ {
		id, err := pm.AllocatePage(types.PageTypeLeaf)
		if err != nil {
			t.Fatalf("AllocatePage() error = %v", err)
		}
		pages = append(pages, id)

		page := &types.Page{
			PageID:   id,
			PageType: types.PageTypeLeaf,
			Records:  []types.Record{{Key: "k", Value: []byte("v")}},
		}
		if err := bp.PutPage(page); err != nil {
			t.Fatalf("PutPage() error = %v", err)
		}
	}

	// Flush all
	if err := bp.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	// Verify pages are persisted by reading directly from file
	for _, pageID := range pages {
		_, err := fm.ReadPage(pageID)
		if err != nil {
			t.Errorf("ReadPage() %d after flush error = %v", pageID, err)
		}
	}
}

// TestFileManager_RootPage tests reading and setting root page
func TestFileManager_RootPage(t *testing.T) {
	tmpFile := "test_root_page.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	// Default root page should be 0
	rootPage := fm.RootPage()
	if rootPage != 0 {
		t.Errorf("RootPage() default = %d, want 0", rootPage)
	}

	// Set root page
	if err := fm.SetRootPage(42); err != nil {
		t.Fatalf("SetRootPage() error = %v", err)
	}

	// Verify it was set
	rootPage = fm.RootPage()
	if rootPage != 42 {
		t.Errorf("RootPage() after set = %d, want 42", rootPage)
	}
}

// TestFileManager_Sync tests syncing data to disk
func TestFileManager_Sync(t *testing.T) {
	tmpFile := "test_sync.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	// Write some header data and sync
	if err := fm.SetTotalPages(100); err != nil {
		t.Fatalf("SetTotalPages() error = %v", err)
	}

	if err := fm.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	// Reopen and verify header was synced
	fm2, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() reopen error = %v", err)
	}
	defer fm2.Close()

	if fm2.TotalPages() != 100 {
		t.Errorf("TotalPages() after sync = %d, want 100", fm2.TotalPages())
	}
}

// TestFileManager_ClearPage tests clearing a page
func TestFileManager_ClearPage(t *testing.T) {
	tmpFile := "test_clear.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	pm := NewPageManager(fm)

	// Allocate and write a page
	id, err := pm.AllocatePage(types.PageTypeLeaf)
	if err != nil {
		t.Fatalf("AllocatePage() error = %v", err)
	}

	page := &types.Page{
		PageID:   id,
		PageType: types.PageTypeLeaf,
		Records:  []types.Record{{Key: "test", Value: []byte("data")}},
	}

	if err := pm.WritePage(page); err != nil {
		t.Fatalf("WritePage() error = %v", err)
	}

	// Clear the page on disk
	if err := fm.ClearPage(id); err != nil {
		t.Fatalf("ClearPage() error = %v", err)
	}

	// ClearPage just zeroes the disk, doesn't sync to buffer,
	// so this is mainly testing it doesn't error
}

// TestFileManager_ResetPages tests resetting all pages
func TestFileManager_ResetPages(t *testing.T) {
	tmpFile := "test_reset_pages.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	pm := NewPageManager(fm)

	// Allocate several pages
	ids := make([]uint32, 0)
	for i := 0; i < 5; i++ {
		id, err := pm.AllocatePage(types.PageTypeLeaf)
		if err != nil {
			t.Fatalf("AllocatePage() error = %v", err)
		}
		ids = append(ids, id)

		page := &types.Page{PageID: id, PageType: types.PageTypeLeaf}
		if err := pm.WritePage(page); err != nil {
			t.Fatalf("WritePage() error = %v", err)
		}
	}

	// Reset all pages
	if err := fm.ResetPages(); err != nil {
		t.Fatalf("ResetPages() error = %v", err)
	}

	// TotalPages should be reset to 0
	if fm.TotalPages() != 0 {
		t.Errorf("TotalPages() after reset = %d, want 0", fm.TotalPages())
	}
}

// TestPageManager_Reset tests resetting the page manager
func TestPageManager_Reset(t *testing.T) {
	pm := NewPageManager(nil)

	// Allocate some pages
	for i := 0; i < 5; i++ {
		_, err := pm.AllocatePage(types.PageTypeLeaf)
		if err != nil {
			t.Fatalf("AllocatePage() error = %v", err)
		}
	}

	// Reset
	if err := pm.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// TotalPages should be 0
	if pm.TotalPages() != 0 {
		t.Errorf("TotalPages() after reset = %d, want 0", pm.TotalPages())
	}
}

// TestPageManager_TotalPages tests getting total page count
func TestPageManager_TotalPages(t *testing.T) {
	pm := NewPageManager(nil)

	if pm.TotalPages() != 0 {
		t.Errorf("TotalPages() initially = %d, want 0", pm.TotalPages())
	}

	// Allocate pages
	for i := 0; i < 10; i++ {
		_, err := pm.AllocatePage(types.PageTypeLeaf)
		if err != nil {
			t.Fatalf("AllocatePage() error = %v", err)
		}
	}

	if pm.TotalPages() != 10 {
		t.Errorf("TotalPages() after 10 allocs = %d, want 10", pm.TotalPages())
	}
}

// TestPageManager_ReadPage tests reading a page
func TestPageManager_ReadPage(t *testing.T) {
	tmpFile := "test_read_page.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	pm := NewPageManager(fm)

	// Allocate and write a page
	id, err := pm.AllocatePage(types.PageTypeLeaf)
	if err != nil {
		t.Fatalf("AllocatePage() error = %v", err)
	}

	page := &types.Page{
		PageID:   id,
		PageType: types.PageTypeLeaf,
		Records:  []types.Record{{Key: "read_test", Value: []byte("read_value")}},
	}

	if err := pm.WritePage(page); err != nil {
		t.Fatalf("WritePage() error = %v", err)
	}

	// Read it back
	got, err := pm.ReadPage(id)
	if err != nil {
		t.Fatalf("ReadPage() error = %v", err)
	}

	if len(got.Records) != 1 {
		t.Fatalf("ReadPage() records = %d, want 1", len(got.Records))
	}

	if got.Records[0].Key != "read_test" {
		t.Errorf("ReadPage() key = %s, want read_test", got.Records[0].Key)
	}
}

// TestFileManager_WritePayload tests writing raw payload
func TestFileManager_WritePayload(t *testing.T) {
	tmpFile := "test_write_payload.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	// Write payload
	testData := []byte("test_payload_data")
	if err := fm.WritePayload(testData); err != nil {
		t.Fatalf("WritePayload() error = %v", err)
	}
}

// TestFileManager_ReadPayload tests reading raw payload
func TestFileManager_ReadPayload(t *testing.T) {
	tmpFile := "test_read_payload.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	// Write and read payload
	testData := []byte("read_payload_test")
	if err := fm.WritePayload(testData); err != nil {
		t.Fatalf("WritePayload() error = %v", err)
	}

	got, err := fm.ReadPayload()
	if err != nil {
		t.Fatalf("ReadPayload() error = %v", err)
	}

	if string(got) != string(testData) {
		t.Errorf("ReadPayload() = %s, want %s", got, testData)
	}
}

// TestPageManager_AllocatePageMultiple tests allocating multiple pages sequentially
func TestPageManager_AllocatePageMultiple(t *testing.T) {
	pm := NewPageManager(nil)

	ids := make([]uint32, 10)
	for i := 0; i < 10; i++ {
		id, err := pm.AllocatePage(types.PageTypeLeaf)
		if err != nil {
			t.Fatalf("AllocatePage() %d error = %v", i, err)
		}
		ids[i] = id

		// IDs should be sequential
		if id != uint32(i) {
			t.Errorf("AllocatePage() %d got ID %d, want %d", i, id, i)
		}
	}
}

// TestPageManager_FreePageReuse tests that freed pages are reused
func TestPageManager_FreePageReuse(t *testing.T) {
	pm := NewPageManager(nil)

	// Allocate and free first page
	id0, err := pm.AllocatePage(types.PageTypeLeaf)
	if err != nil {
		t.Fatalf("AllocatePage() error = %v", err)
	}
	if id0 != 0 {
		t.Errorf("first alloc = %d, want 0", id0)
	}

	if err := pm.FreePage(id0); err != nil {
		t.Fatalf("FreePage() error = %v", err)
	}

	// Allocate second page - should reuse page 0
	id1, err := pm.AllocatePage(types.PageTypeLeaf)
	if err != nil {
		t.Fatalf("AllocatePage() after free error = %v", err)
	}

	if id1 != 0 {
		t.Errorf("after free alloc = %d, want 0 (reuse)", id1)
	}
}

// TestBufferPool_MultipleOperations tests complex buffer pool scenarios
func TestBufferPool_MultipleOperations(t *testing.T) {
	tmpFile := "test_buffer_complex.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	pm := NewPageManager(fm)
	bp := NewBufferPool(3, pm)

	// Allocate 5 pages but buffer only holds 3
	pageIDs := make([]uint32, 5)
	for i := 0; i < 5; i++ {
		id, err := pm.AllocatePage(types.PageTypeLeaf)
		if err != nil {
			t.Fatalf("AllocatePage() error = %v", err)
		}
		pageIDs[i] = id

		page := &types.Page{
			PageID:   id,
			PageType: types.PageTypeLeaf,
			Records:  []types.Record{{Key: "k", Value: []byte(string(rune(i)))}},
		}
		if err := bp.PutPage(page); err != nil {
			t.Fatalf("PutPage() error = %v", err)
		}
	}

	// Buffer should only have 3 pages (LRU evicted first 2)
	if bp.Len() != 3 {
		t.Errorf("buffer pool len = %d, want 3", bp.Len())
	}

	// First two pages should be evicted to disk
	for i := 0; i < 2; i++ {
		_, err := fm.ReadPage(pageIDs[i])
		if err != nil {
			t.Errorf("page %d should be persisted: %v", i, err)
		}
	}

	// Delete a page from buffer
	if err := bp.DeletePage(pageIDs[3]); err != nil {
		t.Fatalf("DeletePage() error = %v", err)
	}

	// Buffer should now have 2 pages
	if bp.Len() != 2 {
		t.Errorf("after delete buffer len = %d, want 2", bp.Len())
	}
}

// TestSerializePage_EdgeCases tests page serialization edge cases
func TestSerializePage_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		page *types.Page
	}{
		{"empty_page", &types.Page{PageID: 0, PageType: types.PageTypeLeaf, Records: []types.Record{}}},
		{"single_small_record", &types.Page{
			PageID:   1,
			PageType: types.PageTypeLeaf,
			Records:  []types.Record{{Key: "k", Value: []byte("v")}},
		}},
		{"multiple_records", &types.Page{
			PageID:   2,
			PageType: types.PageTypeLeaf,
			Records: []types.Record{
				{Key: "k1", Value: []byte("v1"), Deleted: false},
				{Key: "k2", Value: []byte("v2"), Deleted: true},
				{Key: "k3", Value: []byte("v3"), Deleted: false},
			},
		}},
		{"with_next_page", &types.Page{
			PageID:   3,
			PageType: types.PageTypeLeaf,
			NextPage: 99,
			Records:  []types.Record{{Key: "k", Value: []byte("v")}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serialized, err := SerializePage(tt.page)
			if err != nil {
				t.Fatalf("SerializePage() error = %v", err)
			}

			deserialized, err := DeserializePage(serialized)
			if err != nil {
				t.Fatalf("DeserializePage() error = %v", err)
			}

			if deserialized.PageID != tt.page.PageID {
				t.Errorf("PageID mismatch: %d vs %d", deserialized.PageID, tt.page.PageID)
			}

			if deserialized.NextPage != tt.page.NextPage {
				t.Errorf("NextPage mismatch: %d vs %d", deserialized.NextPage, tt.page.NextPage)
			}

			if len(deserialized.Records) != len(tt.page.Records) {
				t.Errorf("Records count mismatch: %d vs %d", len(deserialized.Records), len(tt.page.Records))
			}
		})
	}
}
