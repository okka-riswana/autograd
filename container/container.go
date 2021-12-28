package container

import (
	"os/exec"
)

type Container interface {
	// Command returns a command that will be executed inside the container.
	// It returns nil if the container has been pruned.
	Command(name string, args ...string) *exec.Cmd

	// HostSideSharedDir is the directory in the host side where host and container process
	// can write/read the same data. It will try to return the absolute path of the directory by
	// calling filepath.Abs.
	HostSideSharedDir() string

	// ContainerSideSharedDir is the directory in the container side where host and container process
	// can write/read the same data. It will return the absolute path of the directory.
	ContainerSideSharedDir() string

	// Prune removes the container and its shared directory.
	// This should be called after the container is no longer needed to avoid resource leak.
	Prune() error
}
