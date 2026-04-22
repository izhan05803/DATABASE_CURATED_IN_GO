package engine

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

func TestEngine_ConcurrentSave(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "concur.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent: %v", err)
	}
	defer e.Close()
	var wg sync.WaitGroup
	writers := 8
	saves := 100
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < saves; i++ {
				key := fmt.Sprintf("k%d:%d", id, i)
				if err := e.Set(key, []byte("v")); err != nil {
					t.Errorf("Set err: %v", err)
				}
				if err := e.Save(); err != nil {
					t.Errorf("Save err: %v", err)
					return
				}
			}
		}(w)
	}
	wg.Wait()
	if err := e.Close(); err != nil {
		t.Fatalf("close err: %v", err)
	}
	e2, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("reopen err: %v", err)
	}
	defer e2.Close()
	if _, err := e2.Get("k0:0"); err != nil {
		t.Fatalf("missing key after reopen: %v", err)
	}
}
