// Package ui provides embedded static file serving.
package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed "html" "static/dist"
var Files embed.FS

func staticFiles() fs.FS {
	return Files
}

// Func ServeStaticFiles serves all embeded static files.
func ServeStaticFiles() http.Handler {
	return http.FileServerFS(staticFiles())
}
