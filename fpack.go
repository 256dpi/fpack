// Package fpack provides a functional approach to encoding and decoding byte
// sequences.
package fpack

import "errors"

// ErrBufferTooShort is returned if the provided buffer is too short.
var ErrBufferTooShort = errors.New("buffer too short")
