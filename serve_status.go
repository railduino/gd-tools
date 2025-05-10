package main

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ServeParsePage(w, r, serveStatusContent, serveConfig)
	if err != nil {
		return
	}

	page := ServePage{
		Title:   T("web-status-title"),
		Content: content,
	}

	page.Render(w, r)
}

func BasicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != serveConfig.SysAdmin || !checkPasswordHash(password, serveConfig.Password) {
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
