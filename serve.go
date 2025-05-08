package main

import (
	"context"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

//go:embed www/**
var wwwFS embed.FS

type ServePage struct {
	Title      string
	ProgName   string
	ProgLink   string
	ImprintURL string
	ProtectURL string
	Content    template.HTML
	Request    *http.Request

	ServeConfig
	LoggedIn bool
}

var (
	serveMux    *http.ServeMux
	serveLayout string
)

func RunWebServer(serveAddr string) error {
	serveMux = http.NewServeMux()

	layoutFile := "www/application.html"
	layoutContent, err := wwwFS.ReadFile(layoutFile)
	if err != nil {
		return err
	}
	serveLayout = string(layoutContent)

	subFS, _ := fs.Sub(wwwFS, "www")
	staticServer := http.FileServer(http.FS(subFS))
	serveMux.Handle("/static/", http.StripPrefix("/", staticServer))

	ServeHomeInit()
	ServeStatusInit()

	webServer := &http.Server{
		Handler:      LocaleMiddleware(serveMux),
		Addr:         serveAddr,
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

func (p ServePage) Render(w http.ResponseWriter, r *http.Request) {
	parsedTemplate, err := template.New("app").Funcs(template.FuncMap{
		"T": func(msg string) string {
			return WebT(r, msg)
		},
	}).Parse(serveLayout)
	if err != nil {
		http.Error(w, "Internal error (parse)", 500)
		log.Println("ERROR: parse template:", err)
	}

	if err := parsedTemplate.Execute(w, p); err != nil {
		http.Error(w, "Internal error (execute)", 500)
		log.Println("ERROR: execute template:", err)
	}
}
