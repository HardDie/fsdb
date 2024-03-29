package utils

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

const (
	MaxFilenameLength = 200
)

var (
	reg                = regexp.MustCompile(`[^\p{L}0-9_]+`)
	uniqForbiddenNames = map[string]struct{}{
		// windows
		"con":  {},
		"prn":  {},
		"aux":  {},
		"nul":  {},
		"com1": {},
		"com2": {},
		"com3": {},
		"com4": {},
		"com5": {},
		"com6": {},
		"com7": {},
		"com8": {},
		"com9": {},
		"com0": {},
		"lpt1": {},
		"lpt2": {},
		"lpt3": {},
		"lpt4": {},
		"lpt5": {},
		"lpt6": {},
		"lpt7": {},
		"lpt8": {},
		"lpt9": {},
		"lpt0": {},
	}
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
	if _, ok := uniqForbiddenNames[res]; ok {
		return ""
	}
	return res
}
func Allocate[T any](val T) *T {
	return &val
}

func StructToJSON[T any](val T, isPretty bool) ([]byte, error) {
	var dataJson bytes.Buffer
	enc := json.NewEncoder(&dataJson)
	if isPretty {
		enc.SetIndent("", "\t")
	}
	err := enc.Encode(val)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return dataJson.Bytes(), nil
}
func JSONToStruct[T any](data []byte) (*T, error) {
	var res T
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return &res, nil
}

func Compare[T comparable](a, b *T) bool {
	switch {
	case a == nil && b == nil:
	case a != nil && b != nil:
		if *a != *b {
			return false
		}
	default:
		return false
	}
	return true
}
func CompareSlice[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for _, vala := range a {
		found := false
		for _, valb := range b {
			if vala == valb {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	return true
}
