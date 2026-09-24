package models

import "time"

// Channel & status notifikasi.
const (
	ChannelEmail = "EMAIL"
	ChannelPush  = "PUSH"
	ChannelInApp = "INAPP"

	NotifSent   = "SENT"
	NotifFailed = "FAILED"
)

// NotifTemplate adalah baris tabel CPNOTIFTEMPLATE.
// Body mendukung variabel {{nama}}, {{kode}}, {{role}}, {{detail}}.
type NotifTemplate struct {
	ID        int       `json:"-"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Channel   string    `json:"channel"`
	Subject   string    `json:"subject,omitempty"`
	Body      string    `json:"body"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NotifLog adalah baris tabel CPNOTIFLOG.
type NotifLog struct {
	ID           int       `json:"-"`
	Code         string    `json:"code"`
	TemplateCode string    `json:"template_code,omitempty"`
	Channel      string    `json:"channel"`
	Recipient    string    `json:"recipient"`
	Subject      string    `json:"subject,omitempty"`
	Body         string    `json:"body"`
	Status       string    `json:"status"`
	Error        string    `json:"error,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// NotifSendRequest adalah payload POST /api/notifications/send.
type NotifSendRequest struct {
	TemplateCode string            `json:"template_code"`
	Recipient    string            `json:"recipient"`
	Variables    map[string]string `json:"variables"`
}
