package main

import (
	"bytes"
	"net/http"
	"time"
)

func (app *app) render(w http.ResponseWriter, r *http.Request, status int, templateLayout, templatePage string, data any) {
	ts, ok := app.templateCache[templatePage]
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, templateLayout, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func newTemplateData(r *http.Request) *templateData {
	return &templateData{
		Metadata: metadata{
			Year: time.Now().Year(),
		},
	}
}
