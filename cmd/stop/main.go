package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	pidFile := filepath.Join(os.TempDir(), "golang-backend.pid")
	b, err := os.ReadFile(pidFile)
	if err == nil {
		pidStr := strings.TrimSpace(string(b))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			p, err := os.FindProcess(pid)
			if err == nil {
				if err := p.Signal(os.Interrupt); err == nil {
					fmt.Printf("Sent SIGINT to PID %d, waiting...\n", pid)
					for i := 0; i < 10; i++ {
						if _, err := os.FindProcess(pid); err != nil {
							break
						}
						exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid)).Run()
					}
				}
			}
			exec.Command("taskkill", "/PID", pidStr, "/T", "/F").Run()
			os.Remove(pidFile)
			fmt.Printf("Stopped PID %s\n", pidStr)
			return
		}
	}
	fmt.Println("PID file not found, trying taskkill by name...")
	cmd := exec.Command("taskkill", "/IM", "app.exe", "/T", "/F")
	out, _ := cmd.CombinedOutput()
	fmt.Println(string(out))
	if strings.Contains(string(out), "SUCCESS") {
		return
	}
	cmd2 := exec.Command("taskkill", "/IM", "golang-backend.exe", "/T", "/F")
	out2, _ := cmd2.CombinedOutput()
	fmt.Println(string(out2))
	os.Remove(pidFile)
}
