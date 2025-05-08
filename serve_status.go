package main

import (
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func ServeStatus(w http.ResponseWriter, r *http.Request) {
	page := ServePage{
		Title:      T("web-status-title"),
		ProgName:   "gd-tools",
		ProgLink:   "https://github.com/railduino/gd-tools",
		ImprintURL: serveConfig.ImprintURL,
		ProtectURL: serveConfig.ProtectURL,
		Content: template.HTML(`
<section class="section">
  <h1 class="title">Systemstatus</h1>
  <p>WebSocket wird aufgebaut…</p>
  <div id="status-output"></div>
  <script>
    const socket = new WebSocket("ws://" + location.host + "/ws");
    socket.onmessage = function(event) {
      document.getElementById("status-output").textContent = event.data;
    };
  </script>
</section>
		`),
		Request: r,
	}

	page.Render(w, r)
}

func ServeStatusInit() {
	serveMux.HandleFunc("/status", BasicAuthMiddleware(
		serveConfig.SysAdmin,
		serveConfig.Password,
		ServeStatus,
	))
}

func BasicAuthMiddleware(user, hash string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != user || !checkPasswordHash(password, hash) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
