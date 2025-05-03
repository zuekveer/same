package apperr

import (
	"errors"
)

var (
	ErrAlreadyExist       = errors.New("already exists")
	ErrNotFound           = errors.New("not found")
	ErrConditionViolation = errors.New("condition violation")
	ErrUnauthorized       = errors.New("unauthorized")
)
