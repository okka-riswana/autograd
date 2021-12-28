//go:build linux && builtin_alpine
// +build linux,builtin_alpine

package container

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/fahmifan/autograd/utils"
)

// rootfsEmbed is the rootfs embedded in the binary for the builtin alpine container.
// Contains the following:
//
// - MUSL libc version 1.2.2 (glibc is currently not supported)
//
// - GCC version 10.3.1 configured with "--enable-languages=c,c++,d,objc,go,fortran,ada"
//go:embed alpine.tar.gz
var rootfsEmbed embed.FS

const rootfsTarball string = "alpine.tar.gz"

// NewWithBuiltIn returns a new container with the builtin alpine image.
func NewWithBuiltIn() (Container, error) {
	r, err := rootfsEmbed.Open(rootfsTarball)
	if err != nil {
		return nil, fmt.Errorf("error opening rootfs tarball: %w", err)
	}

	baseDir, err := os.MkdirTemp("", "autograd-")
	if err != nil {
		return nil, fmt.Errorf("error creating rootfs directory: %w", err)
	}

	rootfsDir := filepath.Join(baseDir, "rootfs")
	if err := utils.Untargz(rootfsDir, r); err != nil {
		return nil, fmt.Errorf("error extracting rootfs: %w", err)
	}

	logrus.Debugf("%s rootfs extracted to %s", rootfsTarball, rootfsDir)

	return New(baseDir)
}
