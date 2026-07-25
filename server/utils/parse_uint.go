package utils

import (
	"fmt"
	"strconv"
)

// ParseUint converts a string to a uint. Returns a wrapped error
// with a consistent message on failure.
func ParseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value %q: %w", s, err)
	}
	return uint(v), nil
}
