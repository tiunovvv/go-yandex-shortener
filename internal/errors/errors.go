package errors

import "errors"

var (
	ErrKeyAlreadyExists = errors.New("key already exists")
	ErrURLAlreadySaved  = errors.New("full URL is already saved")
)
