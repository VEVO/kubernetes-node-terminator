package main

import (
	"fmt"
	"strings"
)

// Helper function to build a string of keys in the format
// of key=value, delimited by commas
func keysString(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k, v := range m {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(keys, ",")
}
