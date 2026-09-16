package utils

import "strings"

func TrimPrefixes(s string, prefixes ...string) string {
	trimmed := strings.TrimSpace(s)

	for _, prefix := range prefixes {
		p := strings.TrimSpace(prefix)
		if p == "" {
			continue
		}

		if strings.EqualFold(trimmed[:min(len(p), len(trimmed))], p) {
			rest := trimmed[len(p):]
			rest = strings.TrimLeft(rest, ". ")
			return strings.TrimSpace(rest)
		}
	}

	return trimmed
}

func ToCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '_' || r == '-'
	})

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, "")
}
