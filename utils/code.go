package utils

import (
	"crypto/rand"
	"strings"
)

// Format kode publik: PREFIX-XXXXXXXX (ID numerik tidak diekspos ke luar).
const (
	UserCodePrefix      = "USR-"
	AuditCodePrefix     = "AUD-"
	SyslogCodePrefix    = "SYS-"
	NotifLogPrefix      = "NTF-"
	NotifTemplatePrefix = "NTM-"
	EventCodePrefix     = "EVT-"
	TicketTypePrefix    = "TT-"
	OrderCodePrefix     = "ORD-"
	OrderItemPrefix     = "OI-"
	TicketCodePrefix    = "TIX-"
	TicketQRPrefix      = "QR-"
	RefundCodePrefix    = "RFD-"
	ResvCodePrefix      = "RSV-"
	ScanCodePrefix      = "SCN-"
	codeRandLen         = 8
)

// codeAlphabet menghindari karakter ambigu (0/O, 1/I/L).
var codeAlphabet = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

// GenerateCode membuat kode acak berprefix (crypto/rand).
// Penelepon wajib menangani tabrakan unique (coba lagi).
func GenerateCode(prefix string) (string, error) {
	b := make([]byte, codeRandLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString(prefix)
	for _, v := range b {
		sb.WriteRune(codeAlphabet[int(v)%len(codeAlphabet)])
	}
	return sb.String(), nil
}

// GenerateUserCode membuat kode user acak (USR-XXXXXXXX).
func GenerateUserCode() (string, error) {
	return GenerateCode(UserCodePrefix)
}
