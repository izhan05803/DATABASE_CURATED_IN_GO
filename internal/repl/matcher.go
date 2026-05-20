package repl

// MatchPattern checks if a key matches a glob pattern.
// Supports:
//   - = zero or more characters
//     ? = exactly one character
//     other = literal match
//
// Examples:
//
//	MatchPattern("user:123:profile", "user:*:profile") = true
//	MatchPattern("user:123", "user:?????") = true
//	MatchPattern("user:123", "user:*") = true
//	MatchPattern("admin", "user:*") = false
//
// Algorithm: DP table where dp[i][j] = can key[0..i] match pattern[0..j]
// Time: O(n*m), Space: O(n*m) where n=len(key), m=len(pattern)
func MatchPattern(key, pattern string) bool {
	n, m := len(key), len(pattern)

	// dp[i][j] = does key[0..i-1] match pattern[0..j-1]
	dp := make([][]bool, n+1)
	for i := range dp {
		dp[i] = make([]bool, m+1)
	}

	// Base case: empty key matches empty pattern
	dp[0][0] = true

	// Handle patterns like * or ** that can match empty key
	for j := 1; j <= m; j++ {
		if pattern[j-1] == '*' {
			dp[0][j] = dp[0][j-1]
		}
	}

	// Fill DP table
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if pattern[j-1] == '*' {
				// * can match zero chars (dp[i][j-1]) or one+ chars (dp[i-1][j])
				dp[i][j] = dp[i][j-1] || dp[i-1][j]
			} else if pattern[j-1] == '?' || pattern[j-1] == key[i-1] {
				// ? matches any single char, or literal must match
				dp[i][j] = dp[i-1][j-1]
			}
		}
	}

	return dp[n][m]
}

// matchHelper recursively matches key against pattern.
// keyIdx = current position in key
// patIdx = current position in pattern
func matchHelper(key, pattern string, keyIdx, patIdx int) bool {
	// Base case: both exhausted
	if keyIdx == len(key) && patIdx == len(pattern) {
		return true
	}

	// Pattern exhausted but key has characters
	if patIdx == len(pattern) {
		return keyIdx == len(key)
	}

	// Current pattern character
	current := pattern[patIdx]

	switch current {
	case '*':
		// * matches zero or more characters
		// Strategy: try matching rest of pattern at all possible positions
		// First, try matching * to zero characters (skip *)
		if matchHelper(key, pattern, keyIdx, patIdx+1) {
			return true
		}
		// Then try matching * to one or more characters
		if keyIdx < len(key) {
			return matchHelper(key, pattern, keyIdx+1, patIdx)
		}
		return false

	case '?':
		// ? matches exactly one character
		if keyIdx < len(key) {
			return matchHelper(key, pattern, keyIdx+1, patIdx+1)
		}
		return false

	default:
		// Literal character must match exactly
		if keyIdx < len(key) && key[keyIdx] == current {
			return matchHelper(key, pattern, keyIdx+1, patIdx+1)
		}
		return false
	}
}
