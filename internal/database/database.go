// Package database provides the core data models and operations for the
// curated documentation database.
package database

import (
	"errors"
	"strings"
	"time"
)

// Category represents the type of a documentation entry.
type Category string

const (
	CategoryAPI     Category = "API Reference"
	CategoryGuide   Category = "Developer Guide"
	CategoryTutorial Category = "Tutorial"
	CategorySpec    Category = "Specification"
)

// Entry represents a single curated documentation resource.
type Entry struct {
	ID          int
	Title       string
	URL         string
	Description string
	Category    Category
	Tags        []string
	AddedAt     time.Time
}

// DB holds the in-memory collection of curated entries.
type DB struct {
	entries  []Entry
	nextID   int
}

// New creates a new DB pre-loaded with a set of curated seed entries.
func New() *DB {
	db := &DB{nextID: 1}
	db.seed()
	return db
}

// Add inserts a new entry into the database after basic validation.
func (db *DB) Add(title, url, description string, category Category, tags []string) (Entry, error) {
	if strings.TrimSpace(title) == "" {
		return Entry{}, errors.New("title must not be empty")
	}
	if strings.TrimSpace(url) == "" {
		return Entry{}, errors.New("url must not be empty")
	}
	e := Entry{
		ID:          db.nextID,
		Title:       title,
		URL:         url,
		Description: description,
		Category:    category,
		Tags:        tags,
		AddedAt:     time.Now().UTC(),
	}
	db.entries = append(db.entries, e)
	db.nextID++
	return e, nil
}

// All returns a copy of every entry in the database.
func (db *DB) All() []Entry {
	result := make([]Entry, len(db.entries))
	copy(result, db.entries)
	return result
}

// Search returns entries whose title, description, or tags contain the query
// string (case-insensitive).
func (db *DB) Search(query string) []Entry {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return db.All()
	}
	var results []Entry
	for _, e := range db.entries {
		if strings.Contains(strings.ToLower(e.Title), q) ||
			strings.Contains(strings.ToLower(e.Description), q) ||
			containsTag(e.Tags, q) {
			results = append(results, e)
		}
	}
	return results
}

// FilterByCategory returns all entries belonging to the given category.
func (db *DB) FilterByCategory(category Category) []Entry {
	var results []Entry
	for _, e := range db.entries {
		if e.Category == category {
			results = append(results, e)
		}
	}
	return results
}

// containsTag reports whether any tag in the slice contains the substring s.
func containsTag(tags []string, s string) bool {
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), s) {
			return true
		}
	}
	return false
}

// seed populates the database with a curated set of starter entries.
func (db *DB) seed() {
	entries := []struct {
		title, url, description string
		category                Category
		tags                    []string
	}{
		{
			title:       "Go Standard Library",
			url:         "https://pkg.go.dev/std",
			description: "Official documentation for the Go standard library packages.",
			category:    CategoryAPI,
			tags:        []string{"go", "stdlib", "official"},
		},
		{
			title:       "Effective Go",
			url:         "https://go.dev/doc/effective_go",
			description: "Tips for writing clear, idiomatic Go code from the Go team.",
			category:    CategoryGuide,
			tags:        []string{"go", "best-practices", "official"},
		},
		{
			title:       "A Tour of Go",
			url:         "https://go.dev/tour/",
			description: "Interactive introduction to Go covering all major language features.",
			category:    CategoryTutorial,
			tags:        []string{"go", "beginner", "interactive"},
		},
		{
			title:       "Go Memory Model",
			url:         "https://go.dev/ref/mem",
			description: "Specification of the conditions under which reads of variables in one goroutine can be guaranteed to observe values produced by writes to the same variable in a different goroutine.",
			category:    CategorySpec,
			tags:        []string{"go", "concurrency", "memory"},
		},
		{
			title:       "Go Module Reference",
			url:         "https://go.dev/ref/mod",
			description: "Reference documentation for Go modules, including go.mod syntax and module-aware commands.",
			category:    CategoryAPI,
			tags:        []string{"go", "modules", "dependencies"},
		},
		{
			title:       "RESTful API Design Guidelines",
			url:         "https://restfulapi.net/",
			description: "Comprehensive guide to designing clean and consistent REST APIs.",
			category:    CategoryGuide,
			tags:        []string{"api", "rest", "design"},
		},
		{
			title:       "OpenAPI Specification",
			url:         "https://spec.openapis.org/oas/latest.html",
			description: "The OpenAPI Specification (OAS) defines a standard, language-agnostic interface to HTTP APIs.",
			category:    CategorySpec,
			tags:        []string{"api", "openapi", "specification"},
		},
		{
			title:       "Git Documentation",
			url:         "https://git-scm.com/doc",
			description: "Official reference manual, book, and tutorial resources for Git.",
			category:    CategoryAPI,
			tags:        []string{"git", "version-control", "official"},
		},
	}

	for _, e := range entries {
		db.Add(e.title, e.url, e.description, e.category, e.tags) //nolint:errcheck
	}
}
