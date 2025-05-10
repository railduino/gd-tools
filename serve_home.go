package main

import (
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ServeParsePage(w, r, serveHomeContent, serveConfig)
	if err != nil {
		return
	}

	page := ServePage{
		Title:   T("web-home-title"),
		Content: content,
	}

	page.Render(w, r)
}
