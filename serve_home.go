package main

import (
	"html/template"
	"log"
	"net/http"
)

func (sc *ServeConfig) HomeHandler(w http.ResponseWriter, r *http.Request) {
	page := ServePage{
		Title:   T("web-home-title"),
		Layout:  sc.Layout,
		Content: template.HTML(sc.HomeContent),
		Request: r,
	}

	page.Render(w, r)
}
