package gift

import (
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func NormalizeDedupeKey(name string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMark), norm.NFC)
	normalized, _, err := transform.String(t, name)
	if err != nil {
		normalized = name
	}
	normalized = strings.ToLower(normalized)
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

func isMark(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
