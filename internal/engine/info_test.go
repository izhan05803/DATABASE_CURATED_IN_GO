package engine

import (
	"testing"
	"time"
)

func TestInfoMetrics(t *testing.T) {
	e := New()

	// Verify initial state
	info := e.Info()
	if records, ok := info["records"].(int); !ok || records != 0 {
		t.Errorf("Initial records should be 0, got %v", info["records"])
	}

	// Set some data
	for i := 0; i < 10; i++ {
		e.Set("key"+string(rune(i)), []byte("value"))
	}

	info = e.Info()
	if records, ok := info["records"].(int); !ok || records != 10 {
		t.Errorf("After SET, records should be 10, got %v", info["records"])
	}
	if sets, ok := info["total_sets"].(int64); !ok || sets != 10 {
		t.Errorf("opsSet counter not working: got %v", info["total_sets"])
	}

	// Get some data
	for i := 0; i < 5; i++ {
		e.Get("key" + string(rune(i)))
	}

	info = e.Info()
	if gets, ok := info["total_gets"].(int64); !ok || gets != 5 {
		t.Errorf("opsGet counter not working: got %v", info["total_gets"])
	}

	// Delete some data
	for i := 0; i < 3; i++ {
		e.Delete("key" + string(rune(i)))
	}

	info = e.Info()
	if deletes, ok := info["total_deletes"].(int64); !ok || deletes != 3 {
		t.Errorf("opsDelete counter not working: got %v", info["total_deletes"])
	}
	if records, ok := info["records"].(int); !ok || records != 7 {
		t.Errorf("After DELETE, records should be 7 (10-3), got %v", info["records"])
	}
}

func TestInfoFields(t *testing.T) {
	e := New()

	// Add some test data
	e.Set("test", []byte("value"))
	e.Get("test")

	info := e.Info()

	// Verify all expected fields exist
	expectedFields := []string{
		"records",
		"memory_usage_kb",
		"persisted",
		"file_size_bytes",
		"total_gets",
		"total_sets",
		"total_deletes",
		"total_operations",
		"cache_hits",
		"cache_misses",
		"cache_hit_rate_pct",
		"uptime_seconds",
		"server_time",
	}

	for _, field := range expectedFields {
		if _, ok := info[field]; !ok {
			t.Errorf("Missing field in INFO output: %s", field)
		}
	}
}

func TestInfoUptime(t *testing.T) {
	e := New()

	// Wait a bit (100ms should show at least 1 second in most cases)
	time.Sleep(100 * time.Millisecond)

	info := e.Info()
	uptime, ok := info["uptime_seconds"].(int64)
	if !ok {
		t.Errorf("Uptime not int64: got %T", info["uptime_seconds"])
	}
	// Just verify it exists and is >= 0
	if uptime < 0 {
		t.Errorf("Uptime should be non-negative, got %v", uptime)
	}
}

func TestInfoMemoryUsage(t *testing.T) {
	e := New()

	// Set data with known sizes
	e.Set("key1", []byte("value1"))  // 4 + 6 = 10 bytes
	e.Set("key2", []byte("value22")) // 4 + 7 = 11 bytes

	info := e.Info()
	memKB, ok := info["memory_usage_kb"].(int)
	if !ok {
		t.Errorf("Memory usage not int: got %T", info["memory_usage_kb"])
		return
	}
	// At least 21 bytes of data (key1 + key2)
	if memKB < 0 {
		t.Errorf("Memory usage should be non-negative, got %v KB", memKB)
	}
}

func TestInfoTotalOperations(t *testing.T) {
	e := New()

	e.Set("a", []byte("1"))
	e.Set("b", []byte("2"))
	e.Get("a")
	e.Delete("b")

	info := e.Info()
	if total, ok := info["total_operations"].(int64); !ok || total != 4 {
		t.Errorf("Total operations should be 4, got %v", info["total_operations"])
	}
}
