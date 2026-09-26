// Package pidfile shares the PID-file path and helpers between
// the app and cmd/stop so they can never drift apart.
package pidfile

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const filename = "go-core.pid"

// Path returns the PID file location (OS temp dir, single source of truth).
func Path() string {
	return filepath.Join(os.TempDir(), filename)
}

// legacyPath is the pre-rename name; Read/Remove still handle it as
// fallback so a leftover file from an older build never blocks start/stop.
func legacyPath() string {
	return filepath.Join(os.TempDir(), "golang-backend.pid")
}

// Write stores the current PID with 0600 permissions.
func Write() error {
	return os.WriteFile(Path(), []byte(strconv.Itoa(os.Getpid())), 0600)
}

// Read returns the PID from Path(), falling back to the legacy name.
func Read() (int, error) {
	for _, p := range []string{Path(), legacyPath()} {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
		if err != nil {
			continue
		}
		if pid > 0 {
			return pid, nil
		}
	}
	return 0, os.ErrNotExist
}

// Remove deletes both current and legacy PID files (best-effort).
func Remove() {
	_ = os.Remove(Path())
	_ = os.Remove(legacyPath())
}
