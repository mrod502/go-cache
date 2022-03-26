package gocache

import "errors"

var (
	ErrKey  = errors.New("key not found")
	ErrType = errors.New("TypeError: unable to coerce Key to string")
)
