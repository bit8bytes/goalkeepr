package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"path/filepath"

	"github.com/bit8bytes/goalkeepr/ui"
	"github.com/bit8bytes/goalkeepr/ui/layout"
)

type templateData struct {
	Metadata        metadata
	Form            any
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
	layouts   []layout.Layout
}

// CacheOption configures a templateCache.
type CacheOption func(*templateCache)

// WithFunctions adds custom template functions.
func WithFunctions(fns template.FuncMap) CacheOption {
	return func(tc *templateCache) {
		for key := range fns {
			if _, exists := tc.functions[key]; exists {
				panic(fmt.Sprintf("function %q already exists and cannot be overwritten", key))
			}
		}
		maps.Copy(tc.functions, fns)
	}
}

// newTemplateCache creates a new template cache with optional configuration.
func newTemplateCache(opts ...CacheOption) (map[string]*template.Template, error) {
	tc := &templateCache{
		fsys:      ui.Files,
		functions: defaultFunctions(),
		layouts:   defaultLayouts(),
	}

	for _, opt := range opts {
		opt(tc)
	}

	return tc.build()
}

// defaultFunctions returns the standard template functions.
func defaultFunctions() template.FuncMap {
	return template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"mod": func(a, b int) int { return a % b },
	}
}

// defaultLayouts returns the standard layout configuration.
func defaultLayouts() []layout.Layout {
	return []layout.Layout{
		layout.Convention(layout.Landing),
		layout.Convention(layout.App),
		layout.WithPartials(layout.Settings, layout.App),
		layout.Convention(layout.Share),
		layout.WithPartials(layout.Auth, layout.Share),
		layout.WithoutPartials(layout.Center),
	}
}

// build compiles all templates and returns the cache.
func (tc *templateCache) build() (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	for _, l := range tc.layouts {
		if err := tc.addLayoutPages(cache, l); err != nil {
			return nil, fmt.Errorf("layout %q: %w", l.Name, err)
		}
	}

	return cache, nil
}

// addLayoutPages adds all pages for a given layout to the cache.
func (tc *templateCache) addLayoutPages(cache map[string]*template.Template, l layout.Layout) error {
	for _, pattern := range l.Pages {
		pages, err := fs.Glob(tc.fsys, pattern)
		if err != nil {
			return fmt.Errorf("glob %q: %w", pattern, err)
		}

		for _, page := range pages {
			name := filepath.Base(page)

			patterns := []string{"html/#base.html", l.Layout}
			patterns = append(patterns, l.Partials...)
			patterns = append(patterns, page)

			tmpl, err := template.New(name).Funcs(tc.functions).ParseFS(tc.fsys, patterns...)
			if err != nil {
				return fmt.Errorf("parse template %q: %w", name, err)
			}

			cache[name] = tmpl
		}
	}

	return nil
}
