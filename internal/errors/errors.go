package errors

import "fmt"

type KeyAlreadyExistsError struct {
	Key string
}

func (e *KeyAlreadyExistsError) Error() string {
	return fmt.Sprintf("key %s already exists", e.Key)
}
