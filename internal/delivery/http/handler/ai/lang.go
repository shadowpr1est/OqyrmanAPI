package ai

import "strings"

// resolveLang парсит Accept-Language и возвращает поддерживаемый код языка.
// Поддерживаемые: "ru" (по умолчанию), "kk".
func resolveLang(header string) string {
	if header == "" {
		return "ru"
	}
	// Простой парсер: берём первый тег языка до ',' или ';'.
	tag := header
	if i := strings.IndexAny(tag, ",;"); i >= 0 {
		tag = tag[:i]
	}
	tag = strings.TrimSpace(strings.ToLower(tag))
	if strings.HasPrefix(tag, "kk") {
		return "kk"
	}
	return "ru"
}
