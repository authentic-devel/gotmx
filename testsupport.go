package gotmx

import (
	"strings"
	"testing"
)

func compareStrings(result string, expected string, t *testing.T) {
	differIndex := getDifferIndex(result, expected)

	if differIndex != -1 {

		prefix := result[0:differIndex]

		// We want to make whitespace visible in the log for easier comparing
		result = strings.ReplaceAll(result, " ", ".")
		result = strings.ReplaceAll(result, "\t", ".")
		result = strings.ReplaceAll(result, "\n", "\\n\n")
		expected = strings.ReplaceAll(expected, " ", ".")
		expected = strings.ReplaceAll(expected, "\t", ".")
		expected = strings.ReplaceAll(expected, "\n", "\\n\n")

		t.Errorf("\nExpected result does not match after (index %d):\n"+
			"--------------------------------------------------\n"+
			"%s\n\n"+
			"Expected:\n"+
			"-----------------\n"+
			"%s\n\n"+
			"Actual:\n"+
			"-----------------\n"+
			"%s\n\n",
			differIndex,
			prefix,
			expected,
			result)
	}
}

// write me a function that takes two strings and gets the first index where they differ
func getDifferIndex(a string, b string) int {
	for i := 0; i < len(a); i++ {
		if i >= len(b) {
			return i
		}
		if a[i] != b[i] {
			return i
		}
	}
	if len(a) < len(b) {
		return len(a)
	} else if len(b) < len(a) {
		return len(b)
	}
	return -1
}
