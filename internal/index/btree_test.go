package index

import "testing"

func TestBTree_InsertAndSearch(t *testing.T) {
	tests := []struct {
		name   string
		keys   []string
		search string
		want   uint32
		found  bool
	}{
		{"empty tree", []string{}, "foo", 0, false},
		{"single key found", []string{"foo"}, "foo", 1, true},
		{"single key not found", []string{"foo"}, "bar", 0, false},
		{"multiple keys", []string{"a", "b", "c"}, "b", 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewBTree(3)
			for i, key := range tt.keys {
				tree.Insert(key, uint32(i+1))
			}

			got, found := tree.Search(tt.search)
			if found != tt.found {
				t.Errorf("Search() found = %v, want %v", found, tt.found)
			}
			if found && got != tt.want {
				t.Errorf("Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBTree_Size(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want int
	}{
		{"empty", []string{}, 0},
		{"single", []string{"foo"}, 1},
		{"multiple", []string{"a", "b", "c"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewBTree(3)
			for _, key := range tt.keys {
				tree.Insert(key, 1)
			}
			if got := tree.Size(); got != tt.want {
				t.Errorf("Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBTree_Delete(t *testing.T) {
	tree := NewBTree(3)
	tree.Insert("a", 1)
	tree.Insert("b", 2)
	tree.Insert("c", 3)

	tree.Delete("b")

	if _, found := tree.Search("b"); found {
		t.Error("key 'b' found after delete")
	}

	if tree.Size() != 2 {
		t.Errorf("Size() = %d, want 2", tree.Size())
	}
}

func BenchmarkBTreeInsert(b *testing.B) {
	tree := NewBTree(100)
	for i := 0; i < b.N; i++ {
		tree.Insert(string(rune('a'+i%26)), uint32(i))
	}
}

func BenchmarkBTreeSearch(b *testing.B) {
	tree := NewBTree(100)
	for i := 0; i < 1000; i++ {
		tree.Insert(string(rune('a'+i%26)), uint32(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Search(string(rune('a' + i%26)))
	}
}
