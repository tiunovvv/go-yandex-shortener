package errors

import "errors"

// Errors for shortener.
var (
	ErrKeyAlreadyExists = errors.New("key already exists")
	ErrURLAlreadySaved  = errors.New("full URL already saved")
)
