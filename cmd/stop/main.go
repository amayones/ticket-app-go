package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang-backend/internal/pidfile"
)

func main() {
	if pid, err := pidfile.Read(); err == nil {
		stopByPID(pid)
		return
	}
	fmt.Println("PID file not found, trying fallback by process name...")
	fallbackByName()
}

// stopByPID sends SIGINT (graceful), waits, then force-kills only that PID
// after verifying it still belongs to go-core/app.
func stopByPID(pid int) {
	pidStr := strconv.Itoa(pid)
	p, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("process %d not found: %v\n", pid, err)
		pidfile.Remove()
		return
	}
	// PID reuse guard: refuse to signal a process that is clearly not ours.
	// Indeterminable (tool missing) fails open to avoid bricking stop.
	if name, ok := processName(pid); ok && !isOurBinary(name) {
		fmt.Printf("PID %d belongs to %q, not go-core/app — refusing to kill.\n", pid, name)
		fmt.Println("Delete the stale PID file and try again:")
		fmt.Printf("  del %s\n", pidfile.Path())
		return
	}
	// Graceful: interrupt first.
	if err := p.Signal(os.Interrupt); err != nil {
		fmt.Printf("SIGINT to PID %d failed (%v), will force-kill\n", pid, err)
	} else {
		fmt.Printf("Sent SIGINT to PID %d, waiting up to 8s...\n", pid)
		if waitExit(pid, 8*time.Second) {
			fmt.Printf("Stopped PID %d gracefully\n", pid)
			pidfile.Remove()
			return
		}
		fmt.Printf("PID %d still alive, force-killing...\n", pid)
	}
	forceKill(pidStr)
	pidfile.Remove()
	fmt.Printf("Stopped PID %s\n", pidStr)
}

// waitExit polls liveness: tasklist output on Windows, signal 0 on Unix.
func waitExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !alive(pid) {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return !alive(pid)
}

func alive(pid int) bool {
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").CombinedOutput()
		if err != nil {
			return false
		}
		// tasklist prints "INFO: No tasks..." when PID is gone.
		s := strings.ToUpper(string(out))
		return strings.Contains(s, strconv.Itoa(pid)) && !strings.Contains(s, "INFO:")
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 probes existence on Unix without affecting the process.
	if err := p.Signal(sigZero()); err != nil {
		return false
	}
	return true
}

func forceKill(pidStr string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("taskkill", "/PID", pidStr, "/T", "/F")
	} else {
		cmd = exec.Command("kill", "-9", pidStr)
	}
	out, _ := cmd.CombinedOutput()
	fmt.Println(strings.TrimSpace(string(out)))
}

// processName returns the executable basename for pid.
// ok=false means "could not determine" (caller fails open).
func processName(pid int) (name string, ok bool) {
	pidStr := strconv.Itoa(pid)
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pidStr), "/NH", "/FO", "CSV").CombinedOutput()
		if err != nil {
			return "", false
		}
		// CSV: "Image Name","PID","Session Name",... → first quoted field.
		s := strings.TrimSpace(string(out))
		if !strings.HasPrefix(s, `"`) {
			return "", false
		}
		end := strings.Index(s[1:], `"`)
		if end < 0 {
			return "", false
		}
		return s[1 : 1+end], true
	}
	out, err := exec.Command("ps", "-p", pidStr, "-o", "comm=").CombinedOutput()
	if err != nil {
		return "", false
	}
	name = strings.TrimSpace(string(out))
	if name == "" || strings.HasPrefix(name, "ps ") {
		return "", false
	}
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return name, true
}

// isOurBinary allowlists what stop.exe may signal (exact basename match).
func isOurBinary(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, ok := range []string{"app.exe", "go-core.exe", "golang-backend.exe", "go.exe", "main.exe", "app", "go-core", "golang-backend", "go", "main", "exe"} {
		if n == ok {
			return true
		}
	}
	return false
}

// fallbackByName only runs when no PID file exists. It lists candidates and
// asks for confirmation instead of blindly killing every "app.exe".
func fallbackByName() {
	if runtime.GOOS == "windows" {
		out, _ := exec.Command("tasklist", "/FI", "IMAGENAME eq app.exe", "/NH").CombinedOutput()
		fmt.Println(string(out))
		fmt.Println("Refusing to kill by name automatically: multiple apps may share the name.")
		fmt.Println("Start the app once so the PID file is recreated, then run stop again.")
		return
	}
	out, _ := exec.Command("pgrep", "-af", "go-core|/app").CombinedOutput()
	fmt.Println(string(out))
	fmt.Println("Refusing to kill by name automatically. Use: kill <pid>")
}
