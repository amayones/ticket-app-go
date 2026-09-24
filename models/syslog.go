package models

import "time"

// Level log sistem (CPSYSLOG).
const (
	SyslogError = "ERROR"
	SyslogWarn  = "WARN"
	SyslogInfo  = "INFO"
)

// SysLog adalah baris tabel CPSYSLOG.
type SysLog struct {
	ID        int       `json:"-"`
	Code      string    `json:"code"`
	Level     string    `json:"level"`
	Source    string    `json:"source"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// SyslogFilter menyaring daftar system log.
type SyslogFilter struct {
	Level  string
	Limit  int
	Offset int
}
