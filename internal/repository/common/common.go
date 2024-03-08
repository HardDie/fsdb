package common

import (
	"bytes"
	"encoding/json"

	"github.com/HardDie/fsentry/pkg/fsentry_error"
)

// DataToJSON converts a go object to a json string.
func DataToJSON[T any](data T, isIndent bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if isIndent {
		enc.SetIndent("", "	")
	}
	err := enc.Encode(data)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return buf.Bytes(), nil
}

// JSONToData parses a json string from a byte slice and attempts to fill in a go object.
func JSONToData[T any](data []byte) (*T, error) {
	var res T
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, fsentry_error.Wrap(err, fsentry_error.ErrorInternal)
	}
	return &res, nil
}
