//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func signals() []os.Signal {
	return []os.Signal{os.Interrupt, windows.SIGTERM}
}
