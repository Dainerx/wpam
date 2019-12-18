package stat

import "errors"

var (
	// ErrDataSizeInvalid is returned data size is not valid to run stats on it.
	ErrDataSizeInvalid = errors.New(`Data size is invalid`)
)
