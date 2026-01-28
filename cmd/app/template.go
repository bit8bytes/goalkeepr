package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"time"

	"github.com/bit8bytes/goalkeepr/ui"
	"github.com/bit8bytes/goalkeepr/ui/page"
)

type templateData struct {
	Metadata metadata
	Form     any
	// Data contains page-specific data. Expected types:
	// - SharePageData: for the public share page
	// - GoalsPageData: for the user's goals page
	// - EditGoalPageData: for the edit goal page
	// - ShareGoalsPageData: for the share goals management page
	// - ErrorPageData: for error pages
	// - map[string]any: for settings page (Account, Branding forms)
	Data            any
	IsAuthenticated bool
	Flash           *flash
}

type metadata struct {
	Year int
}

// templateCache builds and caches HTML templates.
type templateCache struct {
	fsys      fs.FS
	functions template.FuncMap
}

// newTemplateCache creates a new template cache with optional configuration.
func newTemplateCache() (map[string]*template.Template, error) {
	tc := &templateCache{
		fsys:      ui.Views(),
		functions: defaultFunctions(),
	}

	return tc.build()
}

// defaultFunctions returns the standard template functions.
func defaultFunctions() template.FuncMap {
	return template.FuncMap{
		"sub":      func(a, b int) int { return a - b },
		"mod":      func(a, b int) int { return a % b },
		"unixTime": func(timestamp int64) time.Time { return time.Unix(timestamp, 0) },
	}
}

// build compiles all templates and returns the cache.
func (tc *templateCache) build() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	for _, p := range page.All() {
		if err := tc.addPage(cache, p); err != nil {
			return nil, fmt.Errorf("page %q: %w", p.Name(), err)
		}
	}

	return cache, nil
}

// addPage adds a page and its associated layout to the cache.
func (tc *templateCache) addPage(cache map[string]*template.Template, page page.Page) error {
	layout := page.Layout()

	// Build pattern list: base -> layout -> partials -> page
	patterns := []string{"base.html", layout.Name()}
	patterns = append(patterns, layout.Partials()...)
	patterns = append(patterns, page.Name())

	tmpl, err := template.New(page.Name()).Funcs(tc.functions).ParseFS(tc.fsys, patterns...)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	cache[page.Name()] = tmpl
	return nil
}
