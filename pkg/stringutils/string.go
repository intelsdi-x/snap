package stringutils

import "fmt"

// GetFirstChar returns the first character from the input string.
func GetFirstChar(s string) string {
	firstChar := ""
	for _, r := range s {
		firstChar = fmt.Sprintf("%c", r)
		break
	}
	return firstChar
}
