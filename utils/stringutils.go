package stringutils

import "strings"

// TrimMargin takes a string as input and removes whitespace before "|" on each line,
// similar to Kotlin's TrimMargin function.
// This makes it easier to format multiline strings in golang code. without having extra whitespace in the actual string.
func TrimMargin(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		indexOfMargin := strings.Index(lines[i], "|")
		if indexOfMargin != -1 {
			lines[i] = lines[i][indexOfMargin+1:]
		}
	}
	return strings.Join(lines, "\n")
}
