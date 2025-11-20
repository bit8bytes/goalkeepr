// Package ui provides embedded static file serving.
package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed "views" "static/dist"
var files embed.FS

func staticFiles() fs.FS {
	return files
}

func Views() fs.FS {
	fs, err := fs.Sub(staticFiles(), "views")
	if err != nil {
		panic(err)
	}
	return fs
}

// Func ServeStaticFiles serves all embeded static files.
func ServeStaticFiles() http.Handler {
	return http.FileServerFS(staticFiles())
}
