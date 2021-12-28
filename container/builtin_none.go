//go:build !linux || !builtin_alpine
// +build !linux !builtin_alpine

package container

import (
	"errors"
)

func NewWithBuiltIn() (Container, error) {
	return nil, errors.New("builtin container is not supported")
}
