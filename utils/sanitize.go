package utils

// Minimal output sanitizer for user-facing content (news body, event
// descriptions). Strips script/style/iframe/object/embed blocks, event
// handler attributes (onclick, ...), and javascript:/data: URLs. Deliberately
// conservative: unknown formatting tags pass through, active content does not.
// JSON APIs never execute markup, but defense in depth for future renderers.
import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	// RE2 has no backreferences: one block pattern per tag instead.
	blockTags   = []string{"script", "style", "iframe", "object", "embed", "link", "meta"}
	rxBlockTags []*regexp.Regexp
	rxLoneTag   = regexp.MustCompile(`(?i)</?(script|style|iframe|object|embed|link|meta)[^>]*>`)
	rxOnAttr    = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	rxJSURL     = regexp.MustCompile(`(?i)(href|src|xlink:href)\s*=\s*("javascript:[^"]*"|'javascript:[^']*'|javascript:[^\s>]+)`)
	rxDataURL   = regexp.MustCompile(`(?i)(href|src)\s*=\s*("data:text/html[^"]*"|'data:text/html[^']*')`)
)

func init() {
	for _, tag := range blockTags {
		rxBlockTags = append(rxBlockTags, regexp.MustCompile(`(?is)<`+tag+`[^>]*>.*?</`+tag+`\s*>`))
	}
}

// SanitizeHTML removes active content from markup, returning safe markup.
func SanitizeHTML(s string) string {
	if s == "" {
		return ""
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	for _, rx := range rxBlockTags {
		s = rx.ReplaceAllString(s, "")
	}
	s = rxLoneTag.ReplaceAllString(s, "")
	s = rxOnAttr.ReplaceAllString(s, "")
	s = rxJSURL.ReplaceAllString(s, `$1="#"`)
	s = rxDataURL.ReplaceAllString(s, `$1="#"`)
	return strings.TrimSpace(s)
}

// HashString returns the hex SHA-256 of s (viewer/IP dedup fingerprints).
func HashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
