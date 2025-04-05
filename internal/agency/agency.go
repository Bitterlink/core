// Package agency Package for helper functions
package agency

import (
	"strconv"
)

// IsNumeric Checks if input s is int or float.
func IsNumeric(s string) bool {
	_, errInt := strconv.Atoi(s)
	if errInt == nil {
		return true
	}

	_, errFloat := strconv.ParseFloat(s, 64)
	return errFloat == nil
}

// Min Helper for safe logging prefix for API key
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
