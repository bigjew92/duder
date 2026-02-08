package rugs

import (
	"html"
	"math/rand"
	"strings"
)

// RandomInRange returns a random integer between min and max (inclusive)
func RandomInRange(min, max int) int {
	if min >= max {
		return min
	}
	return rand.Intn(max-min+1) + min
}

// Clamp constrains a value to a range
func Clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// DecodeHTML decodes HTML entities in a string
func DecodeHTML(s string) string {
	return html.UnescapeString(s)
}

// IsNumeric checks if a string represents a number
func IsNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// ReplaceAll replaces all occurrences of a substring
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// StringContains checks if a slice contains a string
func StringContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
