package engine

import (
	"path/filepath"
	"testing"
)

// TestEngine_EmptyDatabase tests operations on a completely empty database
func TestEngine_EmptyDatabase(t *testing.T) {
	e := New()

	// Get from empty DB
	_, err := e.Get("any_key")
	if err == nil {
		t.Error("Get() from empty DB should return error")
	}

	// Delete from empty DB should not error
	err = e.Delete("any_key")
	if err == nil {
		t.Error("Delete() from empty DB should return error")
	}

	// Keys from empty DB should return empty
	keys := e.Keys("")
	if len(keys) != 0 {
		t.Errorf("Keys() on empty DB returned %d keys, want 0", len(keys))
	}
}

// TestEngine_SingleRecord tests edge case with exactly one record
func TestEngine_SingleRecord(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "single.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Set single record
	if err := e.Set("only_key", []byte("only_value")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Save and verify
	if err := e.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reload and verify
	e2, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() after save error = %v", err)
	}
	defer e2.Close()

	val, err := e2.Get("only_key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(val) != "only_value" {
		t.Errorf("Get() = %s, want only_value", val)
	}
}

// TestEngine_LargeValue tests storing a large value (multi-page)
func TestEngine_LargeValue(t *testing.T) {
	e := New()

	// Create a large value (~100KB)
	largeValue := make([]byte, 100*1024)
	for i := 0; i < len(largeValue); i++ {
		largeValue[i] = byte(i % 256)
	}

	if err := e.Set("large_key", largeValue); err != nil {
		t.Fatalf("Set() large value error = %v", err)
	}

	got, err := e.Get("large_key")
	if err != nil {
		t.Fatalf("Get() large value error = %v", err)
	}

	if len(got) != len(largeValue) {
		t.Errorf("Get() returned %d bytes, want %d", len(got), len(largeValue))
	}

	for i := 0; i < len(got); i++ {
		if got[i] != largeValue[i] {
			t.Errorf("Get() byte mismatch at index %d: got %d, want %d", i, got[i], largeValue[i])
			break
		}
	}
}

// TestEngine_UpdateSameKey tests overwriting the same key multiple times
func TestEngine_UpdateSameKey(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "update.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Update same key 10 times
	for i := 0; i < 10; i++ {
		value := []byte{byte(i)}
		if err := e.Set("update_key", value); err != nil {
			t.Fatalf("Set() iteration %d error = %v", i, err)
		}
		if err := e.Save(); err != nil {
			t.Fatalf("Save() iteration %d error = %v", i, err)
		}
	}

	// Final value should be byte(9)
	got, err := e.Get("update_key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if len(got) != 1 || got[0] != 9 {
		t.Errorf("Get() = %v, want [9]", got)
	}

	// Reload and verify persistence
	e2, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() after updates error = %v", err)
	}
	defer e2.Close()

	got2, err := e2.Get("update_key")
	if err != nil {
		t.Fatalf("Get() after reload error = %v", err)
	}
	if len(got2) != 1 || got2[0] != 9 {
		t.Errorf("Get() after reload = %v, want [9]", got2)
	}
}

// TestEngine_DeleteAndRecreate tests deleting then recreating a key
func TestEngine_DeleteAndRecreate(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "delete_recreate.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Set, delete, set again
	if err := e.Set("key", []byte("v1")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := e.Delete("key"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := e.Set("key", []byte("v2")); err != nil {
		t.Fatalf("Set() after delete error = %v", err)
	}
	if err := e.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	val, err := e.Get("key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(val) != "v2" {
		t.Errorf("Get() = %s, want v2", val)
	}

	// Reload and verify
	e2, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() after delete+create error = %v", err)
	}
	defer e2.Close()

	val2, err := e2.Get("key")
	if err != nil {
		t.Fatalf("Get() after reload error = %v", err)
	}
	if string(val2) != "v2" {
		t.Errorf("Get() after reload = %s, want v2", val2)
	}
}

// TestEngine_EmptyKey tests edge case of empty string key
func TestEngine_EmptyKey(t *testing.T) {
	e := New()

	// Set empty key
	if err := e.Set("", []byte("empty_key_value")); err != nil {
		t.Fatalf("Set() with empty key error = %v", err)
	}

	// Get empty key
	got, err := e.Get("")
	if err != nil {
		t.Fatalf("Get() with empty key error = %v", err)
	}
	if string(got) != "empty_key_value" {
		t.Errorf("Get() empty key = %s, want empty_key_value", got)
	}

	// Delete empty key
	if err := e.Delete(""); err != nil {
		t.Fatalf("Delete() empty key error = %v", err)
	}

	_, err = e.Get("")
	if err == nil {
		t.Error("Get() after delete empty key should return error")
	}
}

// TestEngine_EmptyValue tests edge case of empty value
func TestEngine_EmptyValue(t *testing.T) {
	e := New()

	// Set empty value
	if err := e.Set("empty_val_key", []byte("")); err != nil {
		t.Fatalf("Set() with empty value error = %v", err)
	}

	// Get and verify empty value
	got, err := e.Get("empty_val_key")
	if err != nil {
		t.Fatalf("Get() empty value error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Get() empty value len = %d, want 0", len(got))
	}
}

// TestEngine_SpecialCharactersInKey tests keys with special characters
func TestEngine_SpecialCharactersInKey(t *testing.T) {
	e := New()

	keys := []string{
		"key:with:colons",
		"key/with/slashes",
		"key\\with\\backslashes",
		"key with spaces",
		"key\twith\ttabs",
		"key\nwith\nnewlines",
		"key\"with\"quotes",
		"key'with'apostrophes",
		"key!@#$%^&*()",
		"key™®©",
	}

	for _, key := range keys {
		value := []byte("value_for_" + key)
		if err := e.Set(key, value); err != nil {
			t.Fatalf("Set() key=%q error = %v", key, err)
		}

		got, err := e.Get(key)
		if err != nil {
			t.Fatalf("Get() key=%q error = %v", key, err)
		}
		if string(got) != string(value) {
			t.Errorf("Get() key=%q: got %s, want %s", key, got, value)
		}
	}
}

// TestEngine_KeyPattern tests pattern matching on Keys()
func TestEngine_KeyPattern(t *testing.T) {
	e := New()

	// Set keys with different patterns
	keys := []string{"user:1:name", "user:1:age", "user:2:name", "post:1:title", "post:1:body"}
	for _, key := range keys {
		if err := e.Set(key, []byte("value")); err != nil {
			t.Fatalf("Set() key=%q error = %v", key, err)
		}
	}

	tests := []struct {
		pattern string
		want    int
	}{
		{"user:*", 3},        // user:1:name, user:1:age, user:2:name
		{"user:1:*", 2},      // user:1:name, user:1:age
		{"post:*", 2},        // post:1:title, post:1:body
		{"*", 5},             // all keys
		{"nonexistent:*", 0}, // no match
	}

	for _, tt := range tests {
		got := e.Keys(tt.pattern)
		if len(got) != tt.want {
			t.Errorf("Keys() pattern=%q got %d, want %d", tt.pattern, len(got), tt.want)
		}
	}
}

// TestEngine_SaveMultipleTimes tests repeated saves don't corrupt data
func TestEngine_SaveMultipleTimes(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "multi_save.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Set, save, set more, save again multiple times
	for round := 0; round < 3; round++ {
		for i := 0; i < 5; i++ {
			key := "key_" + string(rune(round*5+i))
			if err := e.Set(key, []byte("value")); err != nil {
				t.Fatalf("Set() round %d key %d error = %v", round, i, err)
			}
		}
		if err := e.Save(); err != nil {
			t.Fatalf("Save() round %d error = %v", round, err)
		}
	}

	// Verify all 15 keys persist
	expected := 15
	got := e.Keys("*")
	if len(got) != expected {
		t.Errorf("Keys() after multi-save got %d, want %d", len(got), expected)
	}
}

// TestEngine_CloseAndReopen tests data survives close/reopen cycle
func TestEngine_CloseAndReopen(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "close_reopen.godb")

	// First instance
	e1, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() instance 1 error = %v", err)
	}
	if err := e1.Set("k1", []byte("v1")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := e1.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := e1.Close(); err != nil {
		t.Fatalf("Close() instance 1 error = %v", err)
	}

	// Second instance
	e2, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() instance 2 error = %v", err)
	}
	defer e2.Close()

	val, err := e2.Get("k1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(val) != "v1" {
		t.Errorf("Get() = %s, want v1", val)
	}

	// Third instance with more data
	if err := e2.Set("k2", []byte("v2")); err != nil {
		t.Fatalf("Set() instance 2 error = %v", err)
	}
	if err := e2.Save(); err != nil {
		t.Fatalf("Save() instance 2 error = %v", err)
	}
	if err := e2.Close(); err != nil {
		t.Fatalf("Close() instance 2 error = %v", err)
	}

	// Fourth instance verifies both keys
	e3, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() instance 3 error = %v", err)
	}
	defer e3.Close()

	val1, err := e3.Get("k1")
	if err != nil {
		t.Fatalf("Get() k1 error = %v", err)
	}
	val2, err := e3.Get("k2")
	if err != nil {
		t.Fatalf("Get() k2 error = %v", err)
	}

	if string(val1) != "v1" || string(val2) != "v2" {
		t.Errorf("Get() values = (%s, %s), want (v1, v2)", val1, val2)
	}
}

// TestEngine_DeleteNonExistentKey tests deleting key that never existed
func TestEngine_DeleteNonExistentKey(t *testing.T) {
	e := New()

	err := e.Delete("never_existed")
	if err == nil {
		t.Error("Delete() nonexistent key should return error")
	}

	// Set and delete once
	if err := e.Set("key", []byte("val")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := e.Delete("key"); err != nil {
		t.Fatalf("Delete() first time error = %v", err)
	}

	// Verify deleted
	_, err = e.Get("key")
	if err == nil {
		t.Error("Get() after delete should return error")
	}

	// Note: Deleting an already-deleted key is a soft-delete (sets Deleted=true again)
	// so it doesn't return an error. This is by design - allows idempotent deletes.
	if err := e.Delete("key"); err != nil {
		t.Fatalf("Delete() second time should work (idempotent): %v", err)
	}
}

// TestEngine_InMemoryOnlyNoSave tests operations without Save()
func TestEngine_InMemoryOnlyNoSave(t *testing.T) {
	e := New()

	if err := e.Set("k1", []byte("v1")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Should work even without Save()
	val, err := e.Get("k1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(val) != "v1" {
		t.Errorf("Get() = %s, want v1", val)
	}

	// Keys() should work
	keys := e.Keys("*")
	if len(keys) != 1 {
		t.Errorf("Keys() got %d keys, want 1", len(keys))
	}
}

// TestEngine_BulkInsert tests inserting many keys with persistence
func TestEngine_BulkInsert(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "bulk.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Insert 500 keys
	for i := 0; i < 500; i++ {
		key := "bulk_" + string(rune(i))
		if err := e.Set(key, []byte{byte(i % 256)}); err != nil {
			t.Fatalf("Set() key %d error = %v", i, err)
		}
	}

	if err := e.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify count
	keys := e.Keys("*")
	if len(keys) != 500 {
		t.Errorf("Keys() count = %d, want 500", len(keys))
	}

	// Reload and verify
	e2, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() after bulk insert error = %v", err)
	}
	defer e2.Close()

	keys2 := e2.Keys("*")
	if len(keys2) != 500 {
		t.Errorf("Keys() after reload count = %d, want 500", len(keys2))
	}
}

// TestEngine_ConsecutiveDeletes tests deleting multiple keys
func TestEngine_ConsecutiveDeletes(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "deletes.godb")
	e, err := NewPersistent(dbPath)
	if err != nil {
		t.Fatalf("NewPersistent() error = %v", err)
	}
	defer e.Close()

	// Create 10 keys
	for i := 0; i < 10; i++ {
		key := "del_" + string(rune(i))
		if err := e.Set(key, []byte("value")); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// Delete first 5
	for i := 0; i < 5; i++ {
		key := "del_" + string(rune(i))
		if err := e.Delete(key); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
	}

	if err := e.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify only 5 keys remain
	keys := e.Keys("*")
	if len(keys) != 5 {
		t.Errorf("Keys() after deletes got %d, want 5", len(keys))
	}
}
