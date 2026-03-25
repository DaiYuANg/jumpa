package utils

import (
	"strings"

	"github.com/samber/lo"
)

// ParseCSVList splits by "," then trims whitespace and drops empty items.
// It always returns a non-nil slice.
func ParseCSVList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	return lo.FilterMap(parts, func(p string, _ int) (string, bool) {
		v := strings.TrimSpace(p)
		return v, v != ""
	})
}

