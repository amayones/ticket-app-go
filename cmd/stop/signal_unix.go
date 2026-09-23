//go:build !windows

package main

import "syscall"

func sigZero() syscall.Signal {
	return syscall.Signal(0)
}
