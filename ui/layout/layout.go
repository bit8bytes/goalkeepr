// Package layout provides constants and builders for all layouts.
package layout

import "fmt"

type Layout struct {
	Name     string
	Layout   string
	Pages    []string
	Partials []string
}

const (
	Auth string = "auth"

	App      string = "app"
	Settings string = "settings"
	Share    string = "share"

	Center string = "center"

	Landing string = "landing"
)

// Convention builds a layout following standard conventions:
// - html/#layouts/{name}.html
// - html/{name}/*.html
// - html/#partials/{name}/*.html
func Convention(name string) Layout {
	return Layout{
		Name:     name,
		Layout:   fmt.Sprintf("html/#layouts/%s.html", name),
		Pages:    []string{fmt.Sprintf("html/%s/*.html", name)},
		Partials: []string{fmt.Sprintf("html/#partials/%s/*.html", name)},
	}
}

// WithPartials creates a layout using a different partials directory.
func WithPartials(name, partialsName string) Layout {
	l := Convention(name)
	l.Partials = []string{fmt.Sprintf("html/#partials/%s/*.html", partialsName)}
	return l
}

// WithoutPartials creates a layout with no partial templates.
func WithoutPartials(name string) Layout {
	l := Convention(name)
	l.Partials = nil
	return l
}
