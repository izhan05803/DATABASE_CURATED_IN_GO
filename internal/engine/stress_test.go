package engine

import (
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEngine_ConcurrentReads tests multiple goroutines reading concurrently
func TestEngine_ConcurrentReads(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "concurrent_reads.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Pre-populate with 100 keys
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("read_key_%d", i)
		if err := e.Set(key, []byte(fmt.Sprintf("value_%d", i))); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}
	if err := e.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Launch 10 goroutines reading concurrently
	var wg sync.WaitGroup
	readers := 10
	readsPerReader := 100
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for i := 0; i < readsPerReader; i++ {
				keyIndex := (readerID*readsPerReader + i) % 100
				key := fmt.Sprintf("read_key_%d", keyIndex)
				val, err := e.Get(key)
				if err != nil {
					t.Errorf("Get() reader %d iteration %d error = %v", readerID, i, err)
					return
				}
				expected := []byte(fmt.Sprintf("value_%d", keyIndex))
				if string(val) != string(expected) {
					t.Errorf("Get() reader %d: got %s, want %s", readerID, val, expected)
					return
				}
			}
		}(r)
	}
	wg.Wait()
}

// TestEngine_ConcurrentWritesDifferentKeys tests concurrent writes to different keys
func TestEngine_ConcurrentWritesDifferentKeys(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "concurrent_writes_diff.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Launch 10 goroutines each writing 100 different keys
	var wg sync.WaitGroup
	writers := 10
	keysPerWriter := 100
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < keysPerWriter; i++ {
				key := fmt.Sprintf("writer_%d_key_%d", writerID, i)
				value := []byte(fmt.Sprintf("value_from_writer_%d_%d", writerID, i))
				if err := e.Set(key, value); err != nil {
					t.Errorf("Set() writer %d error = %v", writerID, err)
					return
				}
			}
		}(w)
	}
	wg.Wait()

	// Verify all keys were written
	keys := e.Keys("*")
	expected := writers * keysPerWriter
	if len(keys) != expected {
		t.Errorf("Keys() got %d, want %d", len(keys), expected)
	}

	// Save and reload to verify persistence
	if err := e.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	e2, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() after reload error = %v", err)
	}
	defer e2.Close()

	keys2 := e2.Keys("*")
	if len(keys2) != expected {
		t.Errorf("Keys() after reload got %d, want %d", len(keys2), expected)
	}
}

// TestEngine_ConcurrentWritesSameKeys tests concurrent writes to same keys (overwrites)
func TestEngine_ConcurrentWritesSameKeys(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "concurrent_writes_same.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Launch 5 goroutines each updating 5 shared keys 100 times
	var wg sync.WaitGroup
	writers := 5
	sharedKeys := 5
	writesPerWriter := 100

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for iter := 0; iter < writesPerWriter; iter++ {
				for k := 0; k < sharedKeys; k++ {
					key := fmt.Sprintf("shared_key_%d", k)
					value := []byte(fmt.Sprintf("writer_%d_iter_%d", writerID, iter))
					if err := e.Set(key, value); err != nil {
						t.Errorf("Set() writer %d error = %v", writerID, err)
						return
					}
				}
			}
		}(w)
	}
	wg.Wait()

	// Should have exactly 5 keys (overwrites only update values)
	keys := e.Keys("*")
	if len(keys) != sharedKeys {
		t.Errorf("Keys() got %d, want %d", len(keys), sharedKeys)
	}
}

// TestEngine_ConcurrentReadWriteMix tests mixed read/write operations
func TestEngine_ConcurrentReadWriteMix(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "concurrent_mix.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Pre-populate with 50 keys
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("mix_key_%d", i)
		if err := e.Set(key, []byte(fmt.Sprintf("value_%d", i))); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// Launch mixed goroutines
	var wg sync.WaitGroup
	operations := 200

	// 3 reader goroutines
	for r := 0; r < 3; r++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for i := 0; i < operations; i++ {
				keyIndex := i % 50
				key := fmt.Sprintf("mix_key_%d", keyIndex)
				_, err := e.Get(key)
				if err != nil {
					// Key might have been deleted, that's OK
				}
			}
		}(r)
	}

	// 2 writer goroutines
	for w := 0; w < 2; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < operations; i++ {
				key := fmt.Sprintf("mix_key_%d", (writerID*operations+i)%100)
				value := []byte(fmt.Sprintf("updated_%d", i))
				if err := e.Set(key, value); err != nil {
					t.Errorf("Set() writer %d error = %v", writerID, err)
					return
				}
			}
		}(w)
	}

	// 1 deleter goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < operations/10; i++ {
			key := fmt.Sprintf("mix_key_%d", i%50)
			_ = e.Delete(key)
		}
	}()

	wg.Wait()
}

// TestEngine_HighThroughputWrites tests rapid sequential writes
func TestEngine_HighThroughputWrites(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "high_throughput.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	start := time.Now()

	// Write 5000 keys as fast as possible
	for i := 0; i < 5000; i++ {
		key := fmt.Sprintf("throughput_%d", i)
		value := make([]byte, 100) // ~100 bytes each
		if err := e.Set(key, value); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	elapsed := time.Since(start)
	t.Logf("5000 writes took %v (%.0f ops/sec)", elapsed, float64(5000)/elapsed.Seconds())

	// Verify all keys
	keys := e.Keys("*")
	if len(keys) != 5000 {
		t.Errorf("Keys() got %d, want 5000", len(keys))
	}
}

// TestEngine_LargeConcurrentSave tests repeated concurrent saves with data changes
func TestEngine_LargeConcurrentSave(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "large_concurrent_save.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Start with 100 keys
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("base_key_%d", i)
		if err := e.Set(key, []byte("base_value")); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	var wg sync.WaitGroup
	var opCount int64

	// 3 writers continuously adding keys
	for w := 0; w < 3; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				key := fmt.Sprintf("writer_%d_key_%d", writerID, i)
				if err := e.Set(key, []byte("value")); err != nil {
					t.Errorf("Set() error = %v", err)
					return
				}
				atomic.AddInt64(&opCount, 1)
			}
		}(w)
	}

	// 2 goroutines repeatedly saving
	for s := 0; s < 2; s++ {
		wg.Add(1)
		go func(saverID int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				if err := e.Save(); err != nil {
					t.Errorf("Save() saver %d error = %v", saverID, err)
					return
				}
			}
		}(s)
	}

	wg.Wait()

	t.Logf("Completed %d operations with concurrent saves", atomic.LoadInt64(&opCount))

	// Verify final state
	keys := e.Keys("*")
	if len(keys) < 100+3*200 {
		t.Errorf("Keys() got %d, want at least %d", len(keys), 100+3*200)
	}
}

// TestEngine_StressDeleteAndRecreate tests heavy delete/recreate cycles
func TestEngine_StressDeleteAndRecreate(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "stress_delete.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Cycle 1: Create 100 keys, delete all, recreate
	for cycle := 0; cycle < 3; cycle++ {
		// Create
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("stress_key_%d", i)
			if err := e.Set(key, []byte("value")); err != nil {
				t.Fatalf("cycle %d Set() error = %v", cycle, err)
			}
		}

		// Verify count
		keys := e.Keys("*")
		if len(keys) != 100 {
			t.Errorf("cycle %d Keys() got %d, want 100", cycle, len(keys))
		}

		// Delete all
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("stress_key_%d", i)
			if err := e.Delete(key); err != nil {
				t.Fatalf("cycle %d Delete() error = %v", cycle, err)
			}
		}

		// Verify all deleted
		keys = e.Keys("*")
		if len(keys) != 0 {
			t.Errorf("cycle %d after delete Keys() got %d, want 0", cycle, len(keys))
		}

		if err := e.Save(); err != nil {
			t.Fatalf("cycle %d Save() error = %v", cycle, err)
		}
	}
}

// TestEngine_LargeValuesConcurrent tests concurrent operations with large values (but fitting in pages)
func TestEngine_LargeValuesConcurrent(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "large_values.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	var wg sync.WaitGroup
	writers := 5
	valuesPerWriter := 20
	valueSize := 1024 // 1KB each (pages are 4KB)

	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < valuesPerWriter; i++ {
				key := fmt.Sprintf("large_%d_%d", writerID, i)
				value := make([]byte, valueSize)
				// Fill with pattern
				for j := 0; j < len(value); j++ {
					value[j] = byte((writerID*valuesPerWriter + i + j) % 256)
				}
				if err := e.Set(key, value); err != nil {
					t.Errorf("Set() writer %d error = %v", writerID, err)
					return
				}
			}
		}(w)
	}

	wg.Wait()

	if err := e.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify total data written
	keys := e.Keys("*")
	expected := writers * valuesPerWriter
	if len(keys) != expected {
		t.Errorf("Keys() got %d, want %d", len(keys), expected)
	}

	// Reload and spot-check a value
	e2, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() after reload error = %v", err)
	}
	defer e2.Close()

	val, err := e2.Get("large_0_0")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(val) != valueSize {
		t.Errorf("Get() value size = %d, want %d", len(val), valueSize)
	}
}

// TestEngine_RapidOpenClose tests rapidly opening and closing databases
func TestEngine_RapidOpenClose(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "rapid_open_close.godb")

	// Create initial database with some data
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	baseKeyCount := 50
	for i := 0; i < baseKeyCount; i++ {
		key := fmt.Sprintf("key_%d", i)
		if err := e.Set(key, []byte("value")); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}
	if err := e.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	e.Close()

	// Rapidly open and close
	for i := 0; i < 10; i++ {
		e, err := NewPersistent(dbPath)
		if err != nil {
			t.Fatalf("iteration %d NewPersistent() error = %v", i, err)
		}

		// Quick verify - should have base keys + any added in previous iterations
		keys := e.Keys("*")
		expectedMin := baseKeyCount + i // i new keys added so far
		if len(keys) < expectedMin {
			t.Errorf("iteration %d Keys() got %d, want at least %d", i, len(keys), expectedMin)
		}

		// Add one key
		key := fmt.Sprintf("new_key_%d", i)
		if err := e.Set(key, []byte("new_value")); err != nil {
			t.Fatalf("iteration %d Set() error = %v", i, err)
		}

		if err := e.Save(); err != nil {
			t.Fatalf("iteration %d Save() error = %v", i, err)
		}

		if err := e.Close(); err != nil {
			t.Fatalf("iteration %d Close() error = %v", i, err)
		}
	}

	// Final verification: should have base + 10 new keys
	e, err = NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("final NewPersistent() error = %v", err)
	}
	defer e.Close()

	keys := e.Keys("*")
	expected := baseKeyCount + 10
	if len(keys) != expected {
		t.Errorf("final Keys() got %d, want %d", len(keys), expected)
	}
}
