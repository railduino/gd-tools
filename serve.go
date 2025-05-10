package main

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

//go:embed www/**
var wwwFS embed.FS

type ServeConfig struct {
	SysAdmin string `json:"sys_admin"`
	Password string `json:"password"`

	Address string `json:"address"`

	ProgName   string `json:"prog_name"`
	ProgLink   string `json:"prog_link"`
	ImprintURL string `json:"imprint_url"`
	ProtectURL string `json:"protect_url"`
}

var (
	serveMux *http.ServeMux

	serveLayoutContent string
	serveHomeContent   string
	serveStatusContent string
)

type ServePage struct {
	Title   string
	Content template.HTML

	ServeConfig
}

func RunWebServer() error {
	serveMux = http.NewServeMux()

	subFS, _ := fs.Sub(wwwFS, "www")
	staticServer := http.FileServer(http.FS(subFS))
	serveMux.Handle("/static/", http.StripPrefix("/", staticServer))

	serveMux.HandleFunc("/", HomeHandler)
	serveMux.HandleFunc("/status", BasicAuthMiddleware(StatusHandler))

	webServer := &http.Server{
		Handler:      LocaleMiddleware(serveMux),
		Addr:         serveConfig.Address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go ListenRoutine(webServer)

	<-ctx.Done()
	log.Printf("INFO: RunWebServer: interrupted")

	if err := webServer.Shutdown(context.TODO()); err != nil {
		log.Printf("WARN: Shutdown: %s", err.Error())
	}

	return nil
}

func ListenRoutine(srv *http.Server) {
	log.Printf("INFO: server listening on address %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("FATAL: ListenRoutine: %s", err.Error())
	}
}

func ServeLoadPage(name string) (string, error) {
	path := filepath.Join("www", name)

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			content, err = wwwFS.ReadFile(path)
			if err != nil {
				return "", err
			}
			log.Println("INFO: from embed:", path)
		} else {
			return "", err // any other error
		}
	} else {
		log.Println("INFO: from file:", path)
	}

	return string(content), nil
}

func ServeParsePage(w http.ResponseWriter, r *http.Request, content string, data interface{}) (template.HTML, error) {
	var contentHTML bytes.Buffer

	contentTmpl, err := template.New("page").Funcs(template.FuncMap{
		"T": func(key string, args ...interface{}) string {
			return WebT(r, key, args...)
		},
	}).Parse(content)

	if err != nil {
		http.Error(w, "Internal error (page-parse)", 500)
		log.Println("ERROR: parse page-template:", err)
		return "", err
	}

	if err := contentTmpl.Execute(&contentHTML, data); err != nil {
		http.Error(w, "Internal error (page-execute)", 500)
		log.Println("ERROR: execute page-template:", err)
		return "", err
	}

	return template.HTML(contentHTML.String()), nil
}

func (p ServePage) Render(w http.ResponseWriter, r *http.Request) {
	var pageHTML bytes.Buffer

	if p.Title != "" {
		p.Title += " - "
	}
	p.ServeConfig = serveConfig

	parsedLayout, err := template.New("app").Funcs(template.FuncMap{
		"T": func(msg string, args ...interface{}) string {
			return WebT(r, msg, args...)
		},
	}).Parse(serveLayoutContent)

	if err != nil {
		http.Error(w, "Internal error (app-parse)", 500)
		log.Println("ERROR: parse app-template:", err)
		return
	}

	if err := parsedLayout.Execute(&pageHTML, p); err != nil {
		http.Error(w, "Internal error (app-execute)", 500)
		log.Println("ERROR: execute app-template:", err)
		return
	}

	w.Write(pageHTML.Bytes())
}
