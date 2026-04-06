package main

import "strings"

// toPascalCase converts SCREAMING_SNAKE_CASE to PascalCase.
// e.g. "HOLDINGS_EXIST" -> "HoldingsExist"
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			b.WriteString(strings.ToLower(p[1:]))
		}
	}
	return b.String()
}
