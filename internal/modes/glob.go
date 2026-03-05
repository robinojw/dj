package modes

import "strings"

// MatchGlob checks if input matches the glob pattern.
// Supports * wildcard only (not full regex).
func MatchGlob(pattern, input string) bool {
	// Exact match if no wildcards
	if !strings.Contains(pattern, "*") {
		return pattern == input
	}

	// Split on * and check parts in order
	parts := strings.Split(pattern, "*")

	// Must start with first part
	if !strings.HasPrefix(input, parts[0]) {
		return false
	}
	input = strings.TrimPrefix(input, parts[0])

	// Check middle parts in order
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(input, parts[i])
		if idx == -1 {
			return false
		}
		input = input[idx+len(parts[i]):]
	}

	// Must end with last part (unless it's empty from trailing *)
	if len(parts) > 1 && parts[len(parts)-1] != "" {
		if !strings.HasSuffix(input, parts[len(parts)-1]) {
			return false
		}
	}

	return true
}
