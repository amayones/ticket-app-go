package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Ticket QR tokens (TICKETING module): signed, non-sequential, verifiable
// offline by any scanner holding QR_SECRET. Separate issuer from session
// tokens so the two can never be confused.
const TicketTokenIssuer = "tikto-ticket"

// GenerateTicketToken signs {ticket, event} with the QR secret (no expiry;
// validity is bounded by event schedule + ticket status in the database).
func GenerateTicketToken(secret, ticketCode, eventCode string) (string, error) {
	if err := ValidateSecret(secret); err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"ticket": ticketCode,
		"event":  eventCode,
		"iss":    TicketTokenIssuer,
		"iat":    time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateTicketToken verifies signature + issuer and returns the claims.
func ValidateTicketToken(tokenStr, secret string) (ticketCode, eventCode string, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(TicketTokenIssuer))
	if err != nil {
		return "", "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid ticket token")
	}
	ticket, _ := claims["ticket"].(string)
	event, _ := claims["event"].(string)
	if ticket == "" || event == "" {
		return "", "", fmt.Errorf("invalid ticket token")
	}
	return ticket, event, nil
}
