package engine

import (
	"testing"
)

func TestKeysPatternMatching(t *testing.T) {
	e := New()

	// Set up test data
	testData := map[string][]byte{
		"user:1:email":     []byte("alice@example.com"),
		"user:1:phone":     []byte("555-1234"),
		"user:2:email":     []byte("bob@example.com"),
		"user:2:phone":     []byte("555-5678"),
		"config:db:host":   []byte("localhost"),
		"config:db:port":   []byte("5432"),
		"config:cache:ttl": []byte("3600"),
		"session:abc123":   []byte("data1"),
		"session:def456":   []byte("data2"),
		"log:2025:05:19":   []byte("error1"),
		"log:2025:05:20":   []byte("error2"),
	}

	for k, v := range testData {
		e.Set(k, v)
	}

	tests := []struct {
		pattern string
		want    []string
		desc    string
	}{
		{"*", []string{
			"config:cache:ttl",
			"config:db:host",
			"config:db:port",
			"log:2025:05:19",
			"log:2025:05:20",
			"session:abc123",
			"session:def456",
			"user:1:email",
			"user:1:phone",
			"user:2:email",
			"user:2:phone",
		}, "wildcard matches all keys"},

		{"user:*:email", []string{
			"user:1:email",
			"user:2:email",
		}, "* in middle matches segments"},

		{"user:1:*", []string{
			"user:1:email",
			"user:1:phone",
		}, "* at end matches anything"},

		{"config:*:*", []string{
			"config:cache:ttl",
			"config:db:host",
			"config:db:port",
		}, "multiple * segments"},

		{"session:*", []string{
			"session:abc123",
			"session:def456",
		}, "session prefix pattern"},

		{"log:*:05:*", []string{
			"log:2025:05:19",
			"log:2025:05:20",
		}, "date-like pattern with * wildcards"},

		{"user:?:email", []string{
			"user:1:email",
			"user:2:email",
		}, "? matches single digit"},

		{"user:?:*", []string{
			"user:1:email",
			"user:1:phone",
			"user:2:email",
			"user:2:phone",
		}, "? and * combined"},

		{"nonexistent:*", []string{}, "no matches returns empty"},

		{"user:1:email", []string{
			"user:1:email",
		}, "exact match"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := e.Keys(tt.pattern)
			if len(got) != len(tt.want) {
				t.Errorf("Keys(%q) returned %d keys, want %d\ngot:  %v\nwant: %v",
					tt.pattern, len(got), len(tt.want), got, tt.want)
				return
			}
			for i, k := range got {
				if k != tt.want[i] {
					t.Errorf("Keys(%q) key %d = %q, want %q",
						tt.pattern, i, k, tt.want[i])
				}
			}
		})
	}
}

// BenchmarkKeysPatternMatching measures KEYS command performance
func BenchmarkKeysPatternMatching(b *testing.B) {
	e := New()

	// Load test data: 1000 keys
	for i := 0; i < 1000; i++ {
		e.Set("user:"+string(rune(i%100))+":data:"+string(rune(i%10)), []byte("value"))
	}

	patterns := []string{
		"user:*:data:*",
		"user:1:*",
		"user:*",
	}

	for _, pattern := range patterns {
		b.Run("pattern="+pattern, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				e.Keys(pattern)
			}
		})
	}
}
