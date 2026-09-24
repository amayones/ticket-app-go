package utils

import (
	"net"
	"strings"
)

// ClientIP mengambil host dari RemoteAddr (tanpa port).
func ClientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return strings.TrimSpace(remoteAddr)
	}
	return host
}
