package usecase

import "errors"

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrValidation        = errors.New("validation failed")
	ErrNotFound          = errors.New("not found")
)
