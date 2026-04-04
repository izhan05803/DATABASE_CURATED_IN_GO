package database

import (
	"testing"
)

func TestAdd(t *testing.T) {
	db := &DB{nextID: 1}

	e, err := db.Add("Go Docs", "https://pkg.go.dev", "Official Go packages", CategoryAPI, []string{"go"})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if e.ID != 1 {
		t.Errorf("expected ID 1, got %d", e.ID)
	}
	if e.Title != "Go Docs" {
		t.Errorf("expected title 'Go Docs', got %q", e.Title)
	}
}

func TestAddValidation(t *testing.T) {
	db := &DB{nextID: 1}

	if _, err := db.Add("", "https://example.com", "", CategoryGuide, nil); err == nil {
		t.Error("expected error for empty title, got nil")
	}
	if _, err := db.Add("Title", "", "", CategoryGuide, nil); err == nil {
		t.Error("expected error for empty url, got nil")
	}
}

func TestAll(t *testing.T) {
	db := New()
	entries := db.All()
	if len(entries) == 0 {
		t.Error("expected seeded entries, got none")
	}
	// Mutating the returned slice should not affect the internal state.
	entries[0].Title = "modified"
	if db.All()[0].Title == "modified" {
		t.Error("All() should return a copy, not a reference")
	}
}

func TestSearch(t *testing.T) {
	db := New()

	results := db.Search("go")
	if len(results) == 0 {
		t.Error("expected results for query 'go', got none")
	}

	results = db.Search("")
	if len(results) != len(db.All()) {
		t.Errorf("empty query should return all entries: got %d, want %d", len(results), len(db.All()))
	}

	results = db.Search("zzznomatch999")
	if len(results) != 0 {
		t.Errorf("expected no results for unmatched query, got %d", len(results))
	}
}

func TestFilterByCategory(t *testing.T) {
	db := New()

	apis := db.FilterByCategory(CategoryAPI)
	for _, e := range apis {
		if e.Category != CategoryAPI {
			t.Errorf("expected category %q, got %q", CategoryAPI, e.Category)
		}
	}
}
