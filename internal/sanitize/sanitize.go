package sanitize

import (
	"html"
	"strings"
)

func Email(e string) string {
	e = strings.TrimSpace(e)
	e = strings.ToLower(e)
	return e
}

func Password(pw string) string {
	return strings.TrimSpace(pw)
}

func Text(t string) string {
	t = strings.TrimSpace(t)
	return html.EscapeString(t)
}

func Date(d string) string {
	return strings.TrimSpace(d)
}

func Bool(v string) string {
	v = strings.TrimSpace(v)
	v = strings.ToLower(v)
	return v
}
