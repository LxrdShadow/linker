package errors

import (
	"errors"
)

var (
	InvalidHeaderSize             = errors.New("invalid header size")
	InvalidChunkSize              = errors.New("invalid chunk size")
)
