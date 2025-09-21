package main

import (
	"io"
	"net/http"
)

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Welcome from Goalkeepr!")
	})

	return mux
}
