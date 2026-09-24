package utils

import (
	"crypto/rand"
	"strings"
)

// Format kode publik: USR-XXXXXXXX (ID numerik tidak diekspos ke luar).
const (
	UserCodePrefix  = "USR-"
	userCodeRandLen = 8
)

// codeAlphabet menghindari karakter ambigu (0/O, 1/I/L).
var codeAlphabet = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

// GenerateUserCode membuat kode user acak (crypto/rand).
// Penelepon wajib menangani tabrakan unique (coba lagi).
func GenerateUserCode() (string, error) {
	b := make([]byte, userCodeRandLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString(UserCodePrefix)
	for _, v := range b {
		sb.WriteRune(codeAlphabet[int(v)%len(codeAlphabet)])
	}
	return sb.String(), nil
}
