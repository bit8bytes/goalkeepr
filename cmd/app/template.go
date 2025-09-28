package main

import (
	"html/template"
	"io/fs"
	"path/filepath"

	"github.com/bit8bytes/goalkeepr/ui"
	"github.com/bit8bytes/goalkeepr/ui/layout"
)

var functions = template.FuncMap{
	"sub": func(a, b int) int { return a - b },
	"mod": func(a, b int) int { return a % b },
}

type cache map[string]*template.Template

type templateData struct {
	Metadata        metadata
	Form            any
	Data            any
	IsAuthenticated bool
}

type metadata struct {
	Year int
}

func newTemplateCache() (cache, error) {
	cache := cache{}

	layouts := []struct {
		name   string
		layout string
		pages  []string
		extras []string
	}{
		{
			name:   layout.Landing,
			layout: "html/layouts/landing.html",
			pages:  []string{"html/pages/landing/*.html"},
			extras: []string{"html/partials/landing/*.html"},
		},
		{
			name:   layout.App,
			layout: "html/layouts/app.html",
			pages:  []string{"html/pages/app/*.html"},
			extras: []string{"html/partials/app/*.html"},
		},
		{
			name:   layout.Settings,
			layout: "html/layouts/settings.html",
			pages:  []string{"html/pages/settings/*.html"},
			extras: []string{"html/partials/app/*.html"},
		},
		{
			name:   layout.Share,
			layout: "html/layouts/share.html",
			pages:  []string{"html/pages/share/*.html"},
			extras: []string{"html/partials/share/*.html"},
		}, {
			name:   layout.Share,
			layout: "html/layouts/share.html",
			pages:  []string{"html/pages/share/*.html"},
			extras: []string{"html/partials/share/*.html"},
		},
		{
			name:   layout.Auth,
			layout: "html/layouts/auth.html",
			pages:  []string{"html/pages/auth/*.html"},
			extras: []string{"html/partials/share/*.html"},
		},
		{
			name:   layout.Center,
			layout: "html/layouts/center.html",
			pages:  []string{"html/pages/center/*.html"},
		},
	}

	for _, l := range layouts {
		if err := addLayoutPages(&cache, l.layout, l.pages, l.extras); err != nil {
			return nil, err
		}
	}

	return cache, nil
}

func addLayoutPages(cache *cache, layout string, pagePatterns []string, extras []string) error {
	var allPages []string

	for _, pattern := range pagePatterns {
		if filepath.Base(pattern) == pattern {
			allPages = append(allPages, pattern)
		} else {
			pages, err := fs.Glob(ui.Files, pattern)
			if err != nil {
				return err
			}
			allPages = append(allPages, pages...)
		}
	}

	for _, page := range allPages {
		name := filepath.Base(page)

		patterns := []string{layout}
		patterns = append(patterns, extras...)
		patterns = append(patterns, page)

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return err
		}

		(*cache)[name] = ts
	}

	return nil
}
