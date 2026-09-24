package models

import "time"

// RefreshToken adalah baris tabel CPREFRESHTOKEN.
// Relasi ke pemilik via UserCode (CODE CPUSER), bukan ID numerik.
type RefreshToken struct {
	ID        int       `json:"-"`
	UserCode  string    `json:"user_code"`
	Token     string    `json:"-"` // stored as SHA-256 hash; never expose
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
