package utils

import (
	"regexp"
	"strings"
)

const (
	MaxFilenameLength = 200
)

var (
	reg = regexp.MustCompile("[^\\p{L}0-9_]+")
)

func NameToID(in string) string {
	// Convert all symbols to lowercase
	lower := strings.ToLower(in)
	// Replace all spaces to underscore symbol
	underscore := strings.ReplaceAll(lower, " ", "_")
	// Keep only letters, numbers and underscore symbols
	res := reg.ReplaceAllString(underscore, "")
	if len(res) > MaxFilenameLength {
		res = res[0:MaxFilenameLength]
	}
	return res
}
func Allocate[T any](val T) *T {
	return &val
}
