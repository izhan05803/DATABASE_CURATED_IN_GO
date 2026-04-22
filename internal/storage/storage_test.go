package storage

import (
	"os"
	"testing"

	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

func TestPageManager_AllocatePage(t *testing.T) {
	pm := NewPageManager(nil)

	id, err := pm.AllocatePage(types.PageTypeLeaf)
	if err != nil {
		t.Fatalf("AllocatePage() error = %v", err)
	}
	if id != 0 {
		t.Errorf("first page ID = %d, want 0", id)
	}

	page, ok := pm.GetPage(id)
	if !ok {
		t.Fatal("page not found after allocation")
	}

	if page.PageType != types.PageTypeLeaf {
		t.Errorf("page type = %v, want PageTypeLeaf", page.PageType)
	}
}

func TestPageManager_FreePage(t *testing.T) {
	pm := NewPageManager(nil)

	id, err := pm.AllocatePage(types.PageTypeLeaf)
	if err != nil {
		t.Fatalf("AllocatePage() error = %v", err)
	}
	if err := pm.FreePage(id); err != nil {
		t.Fatalf("FreePage() error = %v", err)
	}

	_, ok := pm.GetPage(id)
	if ok {
		t.Error("page found after free")
	}
}

func TestSerializeDeserializePage_RoundTrip(t *testing.T) {
	page := &types.Page{
		PageID:   7,
		PageType: types.PageTypeLeaf,
		NextPage: 8,
		Records: []types.Record{
			{Key: "k1", Value: []byte("v1"), Timestamp: 11, Deleted: false},
			{Key: "k2", Value: []byte("v2"), Timestamp: 22, Deleted: true},
		},
	}

	serialized, err := SerializePage(page)
	if err != nil {
		t.Fatalf("SerializePage() error = %v", err)
	}

	decoded, err := DeserializePage(serialized)
	if err != nil {
		t.Fatalf("DeserializePage() error = %v", err)
	}

	if decoded.PageID != page.PageID {
		t.Errorf("PageID = %d, want %d", decoded.PageID, page.PageID)
	}
	if decoded.PageType != page.PageType {
		t.Errorf("PageType = %d, want %d", decoded.PageType, page.PageType)
	}
	if decoded.NextPage != page.NextPage {
		t.Errorf("NextPage = %d, want %d", decoded.NextPage, page.NextPage)
	}
	if len(decoded.Records) != 2 {
		t.Fatalf("records len = %d, want 2", len(decoded.Records))
	}
	if decoded.Records[0].Key != "k1" || string(decoded.Records[0].Value) != "v1" {
		t.Errorf("record[0] mismatch: %+v", decoded.Records[0])
	}
	if decoded.Records[1].Key != "k2" || !decoded.Records[1].Deleted {
		t.Errorf("record[1] mismatch: %+v", decoded.Records[1])
	}
}

func TestFileManager_WriteReadPage(t *testing.T) {
	tmpFile := "test_page_rw.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	page := &types.Page{
		PageID:   0,
		PageType: types.PageTypeLeaf,
		Records: []types.Record{
			{Key: "foo", Value: []byte("bar"), Timestamp: 123},
		},
	}

	if err := fm.WritePage(page); err != nil {
		t.Fatalf("WritePage() error = %v", err)
	}

	got, err := fm.ReadPage(0)
	if err != nil {
		t.Fatalf("ReadPage() error = %v", err)
	}

	if len(got.Records) != 1 {
		t.Fatalf("records len = %d, want 1", len(got.Records))
	}
	if got.Records[0].Key != "foo" || string(got.Records[0].Value) != "bar" {
		t.Errorf("record mismatch: %+v", got.Records[0])
	}
}

func TestBufferPool_EvictsAndPersists(t *testing.T) {
	tmpFile := "test_buffer_pool.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	pm := NewPageManager(fm)
	bp := NewBufferPool(2, pm)

	for i := 0; i < 3; i++ {
		id, err := pm.AllocatePage(types.PageTypeLeaf)
		if err != nil {
			t.Fatalf("AllocatePage() error = %v", err)
		}
		page := &types.Page{PageID: id, PageType: types.PageTypeLeaf, Records: []types.Record{{Key: "k", Value: []byte("v")}}}
		if err := bp.PutPage(page); err != nil {
			t.Fatalf("PutPage() error = %v", err)
		}
	}

	if bp.Len() != 2 {
		t.Errorf("buffer pool len = %d, want 2", bp.Len())
	}

	if _, err := fm.ReadPage(0); err != nil {
		t.Fatalf("expected evicted page to be persisted: %v", err)
	}
}

func TestFileManager_Initialize(t *testing.T) {
	tmpFile := "test_db.godb"
	defer os.Remove(tmpFile)

	fm, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() error = %v", err)
	}
	defer fm.Close()

	if string(fm.header.Magic[:]) != MagicNumber {
		t.Errorf("magic = %s, want %s", fm.header.Magic, MagicNumber)
	}

	if fm.header.Version != Version {
		t.Errorf("version = %d, want %d", fm.header.Version, Version)
	}

	if fm.header.PageSize != PageSize {
		t.Errorf("page size = %d, want %d", fm.header.PageSize, PageSize)
	}
}

func TestFileManager_ReadHeaderFromExistingFile(t *testing.T) {
	tmpFile := "test_existing_db.godb"
	defer os.Remove(tmpFile)

	fm1, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() create error = %v", err)
	}

	fm1.header.TotalPages = 42
	if err := fm1.writeHeader(); err != nil {
		fm1.Close()
		t.Fatalf("writeHeader() error = %v", err)
	}
	if err := fm1.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	fm2, err := NewFileManager(tmpFile)
	if err != nil {
		t.Fatalf("NewFileManager() reopen error = %v", err)
	}
	defer fm2.Close()

	if fm2.header.TotalPages != 42 {
		t.Errorf("total pages = %d, want 42", fm2.header.TotalPages)
	}
}
