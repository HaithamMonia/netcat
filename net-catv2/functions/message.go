package netcat

import "strings"

// Message struct already defined in server.go

// Helper function to check if a string is empty (after trimming whitespace)
func IsEmpty(str string) bool {
	str = strings.TrimSpace(str)
	return str == ""
}
