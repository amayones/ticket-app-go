//go:build windows

package main

import "os"

func sigZero() os.Signal {
	return os.Interrupt
}
