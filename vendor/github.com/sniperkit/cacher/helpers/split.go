package helpers

import (
	"strings"
)

// URLS ARE AN ENTIRELY REASONABLE WAY TO SEND IDS
func GetLastSplit(s string) string {
	bits := strings.Split(s, "/")
	return bits[len(bits)-1]
}

func GetSecondLastSplit(s string) string {
	bits := strings.Split(s, "/")
	return bits[len(bits)-2]
}
