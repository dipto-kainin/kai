package utils

import (
	"strings"
)

// SplitPath splits a URL path into segments, removing empty strings
func SplitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	segments := strings.Split(path, "/")
	result := make([]string, 0, len(segments))
	for _, s := range segments {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// IsParam checks if a path segment is a parameter (starts with :)
func IsParam(segment string) bool {
	return len(segment) > 0 && segment[0] == ':'
}

// ParamName extracts the parameter name from a segment (removes :)
func ParamName(segment string) string {
	if IsParam(segment) {
		return segment[1:]
	}
	return segment
}

// CleanPath normalizes a path by removing duplicate slashes and ensuring it starts with /
func CleanPath(path string) string {
	if path == "" {
		return "/"
	}
	
	// Ensure path starts with /
	if path[0] != '/' {
		path = "/" + path
	}
	
	// Remove duplicate slashes
	var result strings.Builder
	result.Grow(len(path))
	lastWasSlash := false
	
	for _, char := range path {
		if char == '/' {
			if !lastWasSlash {
				result.WriteRune(char)
				lastWasSlash = true
			}
		} else {
			result.WriteRune(char)
			lastWasSlash = false
		}
	}
	
	// Remove trailing slash unless it's root
	p := result.String()
	if len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	
	return p
}

// JoinPath joins path segments with /
func JoinPath(base, path string) string {
	base = CleanPath(base)
	path = CleanPath(path)
	
	if base == "/" {
		return path
	}
	if path == "/" {
		return base
	}
	
	return base + path
}

// MatchPath checks if a request path matches a route pattern
// Returns true and extracted params if matched
func MatchPath(pattern, path string) (bool, map[string]string) {
	patternSegs := SplitPath(pattern)
	pathSegs := SplitPath(path)
	
	// Length must match
	if len(patternSegs) != len(pathSegs) {
		return false, nil
	}
	
	params := make(map[string]string)
	
	for i, patternSeg := range patternSegs {
		if IsParam(patternSeg) {
			// Extract parameter
			params[ParamName(patternSeg)] = pathSegs[i]
		} else if patternSeg != pathSegs[i] {
			// Literal segment must match exactly
			return false, nil
		}
	}
	
	return true, params
}
