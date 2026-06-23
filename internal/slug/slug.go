package slug

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	reNonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	reLeadingTrailing = regexp.MustCompile(`^-+|-+$`)
)

func From(name string) string {
	s := strings.ReplaceAll(name, "đ", "d")
	s = strings.ReplaceAll(s, "Đ", "d")

	t := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)
	normalized, _, _ := transform.String(t, s)

	lower := strings.ToLower(normalized)
	hyphenated := reNonAlphanumeric.ReplaceAllString(lower, "-")
	result := reLeadingTrailing.ReplaceAllString(hyphenated, "")

	return result
}
