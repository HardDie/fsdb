package fsentry_error

import (
	"errors"
	"fmt"
)

var (
	ErrorBadName     = fmt.Errorf("bad name")
	ErrorBadPath     = fmt.Errorf("bad path")
	ErrorExist       = fmt.Errorf("object exist")
	ErrorNotExist    = fmt.Errorf("object not exist")
	ErrorPermissions = fmt.Errorf("not enough permissions")
	ErrorInternal    = fmt.Errorf("internal error")
)

func Wrap(err, localErr error) error {
	return errors.Join(err, localErr)
}
