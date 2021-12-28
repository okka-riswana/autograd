//go:build linux
// +build linux

package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
)

type container struct {
	basedir string
	logger  *logrus.Entry
	pruned  bool

	_ struct{}
}

// New creates a new container. With root directory inside {baseDir}/rootfs and shared data
// directory inside {baseDir}/data.
//
// The caller must have CAP_SYS_ADMIN because it will need mount and pivot_root system call to
// initialize the container.
func New(baseDir string) (Container, error) {
	c := &container{
		basedir: baseDir,
		logger: logrus.WithFields(
			logrus.Fields{
				"container_dir": baseDir,
			},
		),
	}

	c.logger.Debugf("creating container")

	rootfsDir := filepath.Join(baseDir, "rootfs")
	dataDir := filepath.Join(baseDir, "data")

	if _, err := os.Stat(rootfsDir); err != nil {
		return nil, fmt.Errorf("error checking container rootfs dir: %s: %w", rootfsDir, err)
	}

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("error creating container dir: %w", err)
	}

	return c, nil
}

func (c *container) Command(name string, args ...string) *exec.Cmd {
	if c.pruned {
		return nil
	}
	c.logger.Debugf("create command wrapper: %s %v", name, args)

	// NOTE: This will spawn a new child process with os. Args[0] set to a new command that will
	// signal the child process initialization. This will use the in-memory version (proc/self/exe)
	// of the current binary, it is thus safe to delete or replace the on-disk binary (os.Args[0]).
	// We require this mechanism because of the forking limitations of using Go.

	// NOTE: The child process will be executed under new MNT, UTS, IPC, PID, NET,
	// and USER namespace. This will ensure that the container have proper isolation.

	// NOTE: The child process will be executed with the same UID and GID as the current process
	// mapped as root inside the container process.

	// NOTE: In the child process side, os.Args[1] is the base directory of the container.
	// os.Args[2] is the actual command to be executed and os.Args[3:] are its arguments.

	return &exec.Cmd{
		Path: "/proc/self/exe",
		Args: append([]string{containerInitArg0, c.basedir, name}, args...),
		SysProcAttr: &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
			Cloneflags: syscall.CLONE_NEWNS |
				syscall.CLONE_NEWUTS |
				syscall.CLONE_NEWIPC |
				syscall.CLONE_NEWPID |
				syscall.CLONE_NEWNET |
				syscall.CLONE_NEWUSER,
			UidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 0,
					HostID:      os.Getuid(),
					Size:        1,
				},
			},
			GidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 0,
					HostID:      os.Getgid(),
					Size:        1,
				},
			},
		},
	}
}

func (c *container) HostSideSharedDir() string {
	p := filepath.Join(c.basedir, "data")
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

func (c *container) ContainerSideSharedDir() string {
	return containerSharedDir
}

func (c *container) Prune() error {
	if err := os.RemoveAll(c.basedir); err != nil {
		return fmt.Errorf("error removing container directory: %s: %w", c.basedir, err)
	}
	c.pruned = true
	return nil
}
