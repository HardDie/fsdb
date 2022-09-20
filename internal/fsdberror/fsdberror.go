package fsdberror

import (
	"fmt"
)

var (
	ErrorBadFolderName = fmt.Errorf("bad folder name")
	ErrorBadPath       = fmt.Errorf("bad path")
	ErrorExist         = fmt.Errorf("object exist")
	ErrorNotExist      = fmt.Errorf("object not exist")
	ErrorInternal      = fmt.Errorf("internal error")
)

func Wrap(err, localErr error) error {
	return fmt.Errorf("%s: %w", err, localErr)
}
