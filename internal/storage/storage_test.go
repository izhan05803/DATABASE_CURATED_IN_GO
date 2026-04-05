package storage

import (
	"os"
	"testing"

	"github.com/izhan05803/gofromscratchdb/pkg/types"
)

func TestPageManager_AllocatePage(t *testing.T) {
	pm := NewPageManager()

	id := pm.AllocatePage(types.PageTypeLeaf)
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
	pm := NewPageManager()

	id := pm.AllocatePage(types.PageTypeLeaf)
	pm.FreePage(id)

	_, ok := pm.GetPage(id)
	if ok {
		t.Error("page found after free")
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
