// Package fpack provides a functional approach to encoding and decoding byte
// sequences.
package fpack

import "errors"

// ErrBufferTooShort is returned if the provided buffer is too short.
var ErrBufferTooShort = errors.New("buffer too short")

// ErrRemainingBytes is returned if the provided buffer is not fully consumed.
var ErrRemainingBytes = errors.New("remaining bytes")

// ErrEmptyDelimiter is returned if a provided delimiter is empty.
var ErrEmptyDelimiter = errors.New("empty delimiter")

// ErrInvalidSize is returned if a provided number size is invalid.
var ErrInvalidSize = errors.New("invalid size")
