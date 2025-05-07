package main

import (
	_ "bytes"
	"context"
	"embed"
	_ "errors"
	_ "html/template"
	"io/fs"
	"log"
	"net/http"
	_ "net/url"
	_ "os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed www/**
var wwwFS embed.FS

var (
	ServeMux *http.ServeMux
)

func InitServeWeb(serveAddr string) error {
	ServeMux = http.NewServeMux()

	subFS, _ := fs.Sub(wwwFS, "www")
	staticServer := http.FileServer(http.FS(subFS))
	ServeMux.Handle("/static/", http.StripPrefix("/", staticServer))

	app_srv := &http.Server{
		Handler:      ServeLocaleMiddleware(ServeMux),
		Addr:         serveAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go ServeListen(app_srv)

	<-ctx.Done()
	log.Printf("INFO: ListenAndServe: interrupted")

	if err := app_srv.Shutdown(context.TODO()); err != nil {
		log.Printf("WARN: Shutdown: %s", err.Error())
	}
}

func ServeListen(srv *http.Server) {
	log.Printf("INFO: server listening on address %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("FATAL: %s: %s", "ListenAndServe", err.Error())
	}
}

/*
func ServeSuccess(w http.ResponseWriter, r *http.Request, message string) {
	value := url.QueryEscape(T(r, message))
	cookie := &http.Cookie{
		Name:  SuccessCookie,
		Value: value,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
}

func ServeWarning(w http.ResponseWriter, r *http.Request, message string) {
	value := url.QueryEscape(T(r, message))
	cookie := &http.Cookie{
		Name:  WarningCookie,
		Value: value,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
}

func ServeError(w http.ResponseWriter, r *http.Request, err error) {
	value := url.QueryEscape(T(r, err.Error()))
	cookie := &http.Cookie{
		Name:  ErrorCookie,
		Value: value,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
}

func ServeForget(name string) *http.Cookie {
	cookie := &http.Cookie{
		Name:   name,
		Value:  "",
		MaxAge: -1,
	}
	return cookie
}

func (p Page) Serve(w http.ResponseWriter, r *http.Request) {
	var pageHTML bytes.Buffer

	if p.Title != "" {
		p.Title += " - "
	}

	p.Theme = "light"
	if p.Current != nil {
		if p.Current.Refresh {
			accessToken, _ := LoginCreateJWT(p.Current.Email, 15*time.Minute)
			http.SetCookie(w, &http.Cookie{
				Name:     "access_token",
				Value:    accessToken,
				HttpOnly: false,
				Secure:   os.Getenv("JWT_SECRET") != "",
				Path:     "/",
				Expires:  time.Now().Add(15 * time.Minute),
			})
			Info("new AccessToken for %s", p.Current.Email)
		}
		if p.Current.Dark {
			p.Theme = "dark"
		}
	}

	if cookie, err := r.Cookie(SuccessCookie); err == nil {
		message, _ := url.QueryUnescape(cookie.Value)
		p.Success = message
		http.SetCookie(w, ServeForget(SuccessCookie))
	}
	if cookie, err := r.Cookie(WarningCookie); err == nil {
		message, _ := url.QueryUnescape(cookie.Value)
		p.Warning = message
		http.SetCookie(w, ServeForget(WarningCookie))
	}
	if cookie, err := r.Cookie(ErrorCookie); err == nil {
		message, _ := url.QueryUnescape(cookie.Value)
		p.Error = message
		http.SetCookie(w, ServeForget(ErrorCookie))
	}

	var err error
	tmpl, err := template.New("app").Funcs(template.FuncMap{
		"T": func(key string, args ...interface{}) string {
			return T(r, key, args...)
		},
	}).Parse(app_tmpl)
	if err != nil {
		Error("Serve(%s): %s", p.Title, err.Error())
		ServeError(w, r, err)
		return
	}

	if err := tmpl.Execute(&pageHTML, p); err != nil {
		Error("Serve(%s): %s", p.Title, err.Error())
		ServeError(w, r, err)
		return
	}

	w.Write(pageHTML.Bytes())
}

func ServeRender(w http.ResponseWriter, r *http.Request, name string, data interface{}) (template.HTML, error) {
	var contentHTML bytes.Buffer

	contentInput, err := wwwFS.ReadFile("www/" + name + ".html")
	if err != nil {
		Error("ReadFile(%s): %s", name, err.Error())
		http.NotFound(w, r)
		return "", err
	}

	contentTmpl, err := template.New(name).Funcs(template.FuncMap{
		"T": func(key string, args ...interface{}) string {
			return T(r, key, args...)
		},
		"Input": func(kind, name, label string, value interface{}, plus, attr string) template.HTML {
			label = T(r, label)
			return Input(kind, name, label, value, plus, attr)
		},
	}).Parse(string(contentInput))
	if err != nil {
		Error("Parse(%s): %s", name, err.Error())
		ServeError(w, r, err)
		return "", err
	}

	if err := contentTmpl.Execute(&contentHTML, data); err != nil {
		Error("Execute(%s): %s", name, err.Error())
		ServeError(w, r, err)
		return "", err
	}

	return template.HTML(contentHTML.String()), nil
}

func ServeParse(w http.ResponseWriter, r *http.Request, current *User) error {
	if err := r.ParseForm(); err != nil {
		Error("ParseForm: %s", err.Error())
		ServeError(w, r, err)
		http.Redirect(w, r, "/", http.StatusFound)
		return err
	}

	if ok := current.CheckCSRFToken(r.FormValue("csrf_token")); !ok {
		err := errors.New(T(r, "WrongToken"))
		ServeWarning(w, r, "WrongToken")
		http.Redirect(w, r, "/", http.StatusFound)
		return err
	}

	return nil
}
*/
