package repl

import "testing"

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		key     string
		pattern string
		want    bool
		desc    string
	}{
		// Literal matches
		{"user", "user", true, "exact match"},
		{"user", "admin", false, "exact mismatch"},
		{"", "", true, "empty key and pattern"},

		// Wildcard * (zero or more)
		{"user:123:profile", "user:*:profile", true, "* matches middle segment"},
		{"user::profile", "user:*:profile", true, "* matches zero characters (empty string)"},
		{"user:very:long:nested:profile", "user:*:profile", true, "* matches multiple segments with colons"},
		{"user:123", "user:*", true, "* at end matches anything"},
		{"user", "user:*", false, "missing separator (requires colon)"},
		{"admin:123:profile", "user:*:profile", false, "prefix mismatch"},

		// Wildcard ? (exactly one)
		{"user:abc", "user:???", true, "? matches exactly 3 chars"},
		{"user:ab", "user:???", false, "? requires exact length"},
		{"user:abcd", "user:???", false, "? requires exact length"},
		{"a", "?", true, "? matches single char"},
		{"ab", "?", false, "? doesn't match multiple"},

		// Combined * and ?
		{"user:123:profile", "user:*:profile", true, "* with prefix/suffix"},
		{"user:123:profile", "user:???:*", true, "multiple ?"},
		{"user:abc:profile", "user:???:*", true, "multiple ? match segment"},
		{"user:ab:profile", "user:???:*", false, "? count mismatch"},

		// Edge cases
		{"", "*", true, "* matches empty string"},
		{"", "?", false, "? doesn't match empty"},
		{"a", "*", true, "* matches single char"},
		{"***", "*", true, "literal * in key matches wildcard"},
		{"test*key", "test*key", true, "literal * in both"},

		// Real-world patterns
		{"app:user:123:email", "app:*:*", true, "multiple * segments"},
		{"app:user:123:email", "app:user:*", true, "prefix with wildcard"},
		{"config:db:connection", "config:*:*", true, "nested namespace"},
		{"session:abc123def456", "session:*", true, "session ID pattern"},
		{"log:2025:05:19:error", "log:*:*:*:*", true, "date-like pattern"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := MatchPattern(tt.key, tt.pattern)
			if got != tt.want {
				t.Errorf("MatchPattern(%q, %q) = %v, want %v",
					tt.key, tt.pattern, got, tt.want)
			}
		})
	}
}

// BenchmarkMatchPattern measures performance of pattern matching
func BenchmarkMatchPattern(b *testing.B) {
	tests := []struct {
		name    string
		key     string
		pattern string
	}{
		{"simple_literal", "user:123", "user:123"},
		{"single_wildcard", "user:123:profile", "user:*:profile"},
		{"multiple_wildcard", "app:user:123:email:data", "app:*:*:*"},
		{"complex_pattern", "app:service:v1:endpoint:logs:2025", "app:*:v?:*:*:*"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				MatchPattern(tt.key, tt.pattern)
			}
		})
	}
}
