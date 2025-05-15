package service

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrSlugExists       = errors.New("slug already exists")
	ErrCategoryNotFound = errors.New("category not found")
	ErrInvalidStatus    = errors.New("invalid status")
	ErrInvalidOperation = errors.New("invalid operation")
)
