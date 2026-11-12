package pihole

import "strings"

// escapePathSegment percent-encodes every byte except the RFC 3986 unreserved
// characters. url.PathEscape leaves reserved characters such as '+', '$' and
// '&' literal, and Pi-hole's API does not match those inside a path parameter:
// a DELETE for a regex domain would return 404 while the entry stayed in the
// gravity database.
func escapePathSegment(s string) string {
	const hex = "0123456789ABCDEF"

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' ||
			c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&0x0f])
	}
	return b.String()
}
