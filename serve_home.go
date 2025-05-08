package main

import (
	"html/template"
	"net/http"
)

func ServeHome(w http.ResponseWriter, r *http.Request) {
	page := ServePage{
		Title:      T("web-home-title"),
		ProgName:   "gd-tools",
		ProgLink:   "https://github.com/railduino/gd-tools",
		ImprintURL: serveConfig.ImprintURL,
		ProtectURL: serveConfig.ProtectURL,
		Content: template.HTML(`
<section class="section">
  <div class="content">
    <h1>Willkommen bei gd-tools</h1>
    <p>Diese Seite ist Teil eines internen Systems zur Verwaltung von Serverdiensten.</p>
    <p><strong>Hinweis:</strong> Kein Zugriff auf Systemstatus ohne Autorisierung.</p>
    <p>
      <a href="/status" class="button is-link">Status (Login erforderlich)</a>
    </p>
  </div>
</section>
		`),
		Request: r,
	}

	page.Render(w, r)
}

func ServeHomeInit() {
	serveMux.HandleFunc("/", ServeHome)
}
