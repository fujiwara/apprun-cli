//go:build !windows

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func signals() []os.Signal {
	return []os.Signal{os.Interrupt, unix.SIGTERM}
}
