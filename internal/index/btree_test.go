package index

import (
	"fmt"
	"testing"
)

func TestBTree_InsertAndSearch(t *testing.T) {
	tree := NewBTree(4)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("k%03d", i)
		if err := tree.Insert(key, uint32(i)); err != nil {
			t.Fatalf("Insert(%q) error = %v", key, err)
		}
	}

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("k%03d", i)
		got, found := tree.Search(key)
		if !found {
			t.Fatalf("Search(%q) not found", key)
		}
		if got != uint32(i) {
			t.Fatalf("Search(%q) = %d, want %d", key, got, i)
		}
	}

	if _, found := tree.Search("k999"); found {
		t.Fatal("unexpected key found")
	}
}

func TestBTree_UpdateExistingKey(t *testing.T) {
	tree := NewBTree(4)

	if err := tree.Insert("alpha", 1); err != nil {
		t.Fatalf("Insert error = %v", err)
	}
	if err := tree.Insert("alpha", 99); err != nil {
		t.Fatalf("Insert update error = %v", err)
	}

	got, found := tree.Search("alpha")
	if !found {
		t.Fatal("alpha not found")
	}
	if got != 99 {
		t.Fatalf("Search(alpha) = %d, want 99", got)
	}

	if tree.Size() != 1 {
		t.Fatalf("Size() = %d, want 1", tree.Size())
	}
}

func TestBTree_DeleteAcrossTree(t *testing.T) {
	tree := NewBTree(5)

	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("key-%03d", i)
		if err := tree.Insert(key, uint32(i)); err != nil {
			t.Fatalf("Insert(%q) error = %v", key, err)
		}
	}

	for i := 0; i < 200; i += 2 {
		key := fmt.Sprintf("key-%03d", i)
		if err := tree.Delete(key); err != nil {
			t.Fatalf("Delete(%q) error = %v", key, err)
		}
	}

	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("key-%03d", i)
		_, found := tree.Search(key)
		if i%2 == 0 && found {
			t.Fatalf("expected %q deleted", key)
		}
		if i%2 == 1 && !found {
			t.Fatalf("expected %q present", key)
		}
	}

	if tree.Size() != 100 {
		t.Fatalf("Size() = %d, want 100", tree.Size())
	}
}

func TestBTree_DeleteMissingKey(t *testing.T) {
	tree := NewBTree(4)
	if err := tree.Insert("a", 1); err != nil {
		t.Fatalf("Insert error = %v", err)
	}

	if err := tree.Delete("missing"); err != nil {
		t.Fatalf("Delete missing error = %v", err)
	}

	if tree.Size() != 1 {
		t.Fatalf("Size() = %d, want 1", tree.Size())
	}
}

func TestBTree_SequentialDeletesToEmpty(t *testing.T) {
	tree := NewBTree(4)

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("k%02d", i)
		if err := tree.Insert(key, uint32(i)); err != nil {
			t.Fatalf("Insert(%q) error = %v", key, err)
		}
	}

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("k%02d", i)
		if err := tree.Delete(key); err != nil {
			t.Fatalf("Delete(%q) error = %v", key, err)
		}
	}

	if tree.Size() != 0 {
		t.Fatalf("Size() = %d, want 0", tree.Size())
	}

	if _, found := tree.Search("k00"); found {
		t.Fatal("unexpected key found in empty tree")
	}
}

func BenchmarkBTreeInsert(b *testing.B) {
	tree := NewBTree(16)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench-key-%09d", i)
		_ = tree.Insert(key, uint32(i))
	}
}

func BenchmarkBTreeSearch(b *testing.B) {
	tree := NewBTree(16)
	const count = 100000

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("bench-key-%09d", i)
		_ = tree.Insert(key, uint32(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench-key-%09d", i%count)
		_, _ = tree.Search(key)
	}
}

func BenchmarkMapSearch(b *testing.B) {
	const count = 100000
	m := make(map[string]uint32, count)

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("bench-key-%09d", i)
		m[key] = uint32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench-key-%09d", i%count)
		_ = m[key]
	}
}
