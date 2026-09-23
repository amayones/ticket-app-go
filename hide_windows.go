//go:build windows

package main

import "syscall"

var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	user32         = syscall.NewLazyDLL("user32.dll")
	procGetConsole = kernel32.NewProc("GetConsoleWindow")
	procShowWindow = user32.NewProc("ShowWindow")
)

func hideConsole() {
	hwnd, _, _ := procGetConsole.Call()
	if hwnd != 0 {
		procShowWindow.Call(hwnd, uintptr(0))
	}
}

func showConsole() {
	hwnd, _, _ := procGetConsole.Call()
	if hwnd != 0 {
		procShowWindow.Call(hwnd, uintptr(5))
	}
}
