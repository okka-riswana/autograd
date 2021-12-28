//go:build linux
// +build linux

package container

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/sirupsen/logrus"
)

//containerInitArg0 is the first argument of the container's init process.
const containerInitArg0 = "init-autograd-container"

// containerSharedDir is the directory in the container side where host and container process
// can write/read the same data.
const containerSharedDir = "/mnt/data"

func init() {
	if os.Args[0] == containerInitArg0 {
		logrus.SetFormatter(&logrus.JSONFormatter{DisableHTMLEscape: true, PrettyPrint: false})
		logrus.SetOutput(os.Stderr)
		initMounts()
		runCommand()
		os.Exit(0)
	}
}

func initMounts() {
	// NOTE: os.Args[1] is the base directory of the container.
	// It must contain rootfs/ for the container's root and data/ directory for the container's
	// shared directory.

	// sanity check
	if os.Args[0] != containerInitArg0 {
		return
	}

	baseDir := os.Args[1]
	hostDataDir := path.Join(baseDir, "data")

	newroot := path.Join(baseDir, "rootfs")
	putold := path.Join(newroot, "oldroot")

	// dataDir is the host side data directory to be mounted into the container.
	dataDir := path.Join(newroot, "mnt", "data")

	// procDir is the container process's /proc directory in the host side.
	procDir := path.Join(newroot, "proc")

	// Create and mount the data directory.

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logrus.Fatalf("failed to create data directory: %v\n", err)
	}

	if err := syscall.Mount(
		hostDataDir,
		dataDir,
		"",
		syscall.MS_BIND|syscall.MS_PRIVATE,
		"",
	); err != nil {
		logrus.Fatalf("failed to mount data directory: %s -> %s: %v\n", hostDataDir, dataDir, err)
	}

	if err := syscall.Mount("proc", procDir, "proc", 0, ""); err != nil {
		logrus.Fatalf("failed to mount proc: %v\n", err)
	}

	// Bind mount the new root directory to itself.
	// This is a workaround for the requirement of pivot_root:
	//
	// -  new_root  must  be  a path to a mount point, but can't be "/".  A path that is not
	//    already a mount point can be converted into one by bind mounting the path onto itself.
	//
	// See pivot_root(2) for details.

	if err := syscall.Mount(
		newroot,
		newroot,
		"",
		syscall.MS_BIND|syscall.MS_REC|syscall.MS_PRIVATE,
		"",
	); err != nil {
		logrus.Fatalf("failed to bind mount new rootfs: %v\n", err)
	}

	// Create a new directory for the old root to be pivoted into.
	if err := os.MkdirAll(putold, 0700); err != nil && !os.IsExist(err) {
		logrus.Fatalf("failed create old root dir: %v\n", err)
	}

	// Pivot into the new root.
	if err := syscall.PivotRoot(newroot, putold); err != nil {
		logrus.Fatalf("failed to pivot root: %v\n", err)
	}

	// Change to the new root.
	if err := os.Chdir("/"); err != nil {
		logrus.Fatalf("failed to chdir to /: %v\n", err)
	}

	// old root directory is now in /oldroot

	// Unmount the old root.
	if err := syscall.Unmount("/oldroot", syscall.MNT_DETACH); err != nil {
		logrus.Fatalf("failed to detach old root: %v\n", err)
	}

	// Remove the (now empty) old root.
	if err := os.Remove("/oldroot"); err != nil {
		logrus.Fatalf("failed to remove old root directory: %v\n", err)
	}
}

func runCommand() {
	// NOTE: os.Args[2] is the command to run. os.Args[3:] are its arguments.

	// sanity check
	if os.Args[0] != containerInitArg0 || len(os.Args) < 3 {
		return
	}

	command := os.Args[2]
	var commandArgs []string

	if len(os.Args) > 3 {
		commandArgs = os.Args[3:]
	}

	cmd := exec.Command(command, commandArgs...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{
		"PS1=[ns] # ",
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	}

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			logrus.Errorf("error running the command: %s %v :%v\n", command, commandArgs, err)
		}
	}
}
