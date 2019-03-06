package engine

import "errors"

var (
	// ErrUsernameTaken is whether the desired username is taken on the store.
	ErrUsernameTaken = errors.New("username is already in use")
	// ErrEmailTaken is whether the desired username is taken on the store.
	ErrEmailTaken = errors.New("email is already in use")
	// ErrMissingFields means that required fields are missing.
	ErrMissingFields = errors.New("required fields are missing")
)
