package main

import (
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (sc *ServeConfig) StatusHandler(w http.ResponseWriter, r *http.Request) {
	page := ServePage{
		Title:   T("web-status-title"),
		Layout:  sc.Layout,
		Content: template.HTML(sc.StatusContent),
		Request: r,
	}

	page.Render(w, r)
}

func (sc *ServeConfig) ServeStatusInit() error {
	serveMux.HandleFunc("/status", BasicAuthMiddleware(ServeStatus))
}

func (sc *ServeConfig) BasicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || sc.SysAdmin != user || !checkPasswordHash(password, sc.Password) {
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
