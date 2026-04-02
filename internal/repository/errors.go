package repository

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrInternal = errors.New("internal server error")
)
