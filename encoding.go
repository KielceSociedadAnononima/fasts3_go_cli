package fasts3

import (
	"errors"
	"fmt"
	"strings"
)

// encodePathSegment percent-encodes a single URL path segment according to RFC 3986 unreserved characters.
func encodePathSegment(segment string) string {
	var b strings.Builder
	b.Grow(len(segment))
	for i := 0; i < len(segment); i++ {
		c := segment[i]
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hexUpper(c >> 4))
			b.WriteByte(hexUpper(c & 0x0F))
		}
	}
	return b.String()
}

// encodeObjectKey validates and percent-encodes a multi-segment S3 object key while preserving '/' separators.
func encodeObjectKey(key string) (string, error) {
	normalized := strings.ReplaceAll(key, "\\", "/")
	rawParts := strings.Split(normalized, "/")
	encodedParts := make([]string, 0, len(rawParts))

	for _, part := range rawParts {
		if part == "" {
			continue
		}
		if part == "." || part == ".." {
			return "", fmt.Errorf("fasts3: object key cannot contain relative traversal segment %q", part)
		}
		encodedParts = append(encodedParts, encodePathSegment(part))
	}

	if len(encodedParts) == 0 {
		return "", errors.New("fasts3: object key cannot be empty")
	}
	return strings.Join(encodedParts, "/"), nil
}

// parseContentDispositionFilename extracts `filename="..."` from a Content-Disposition header.
func parseContentDispositionFilename(headerVal string) string {
	for _, part := range strings.Split(headerVal, ";") {
		trimmed := strings.TrimSpace(part)
		if rest, ok := strings.CutPrefix(trimmed, "filename="); ok {
			unquoted := strings.Trim(strings.TrimSpace(rest), `"`)
			if unquoted != "" {
				return unquoted
			}
		}
	}
	return ""
}

func hexUpper(nibble byte) byte {
	if nibble < 10 {
		return '0' + nibble
	}
	return 'A' + (nibble - 10)
}
