package glob

import (
	"path/filepath"
	"strings"
)

// Match checks if a path matches a glob pattern.
// Supports standard glob patterns (* and ?) and ** for recursive directory matching.
func Match(pattern, path string) bool {
	if strings.HasPrefix(pattern, "!") {
		return false
	}
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)
	if strings.Contains(pattern, "**") {
		base := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(path, base+"/") || path == base
	}
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false
	}
	return matched
}

// MatchAny returns true if the path matches any of the provided patterns.
func MatchAny(patterns []string, path string) bool {
	for _, pattern := range patterns {
		if Match(pattern, path) {
			return true
		}
	}
	return false
}
