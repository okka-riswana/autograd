//go:build !linux
// +build !linux

package container

import (
	"errors"
)

func New(basedir string) (Container, error) {
	return nil, errors.New("container is not supported on this platform")
}
