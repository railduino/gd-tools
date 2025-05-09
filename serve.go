package main

import (
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

	LayoutContent []byte `json:"-"`
	HomeContent   []byte `json:"-"`
	StatusContent []byte `json:"-"`

	Mux *http.ServeMux `json:"-"`
}

type ServePage struct {
	Title    string
	Layout   string
	Content  template.HTML
	Request  *http.Request
	LoggedIn bool
}

func (sc *ServeConfig) RunWebServer() error {
	subFS, _ := fs.Sub(wwwFS, "www")
	staticServer := http.FileServer(http.FS(subFS))
	sc.Mux.Handle("/static/", http.StripPrefix("/", staticServer))

	webServer := &http.Server{
		Handler:      LocaleMiddleware(sc.Mux),
		Addr:         sc.Address,
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

func ServeLoadPage(name string) ([]byte, error) {
	path := filepath.Join("www", name)

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			content, err = wwwFS.ReadFile(path)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err // any other error
		}
	}

	return content, nil
}

func (p ServePage) Render(w http.ResponseWriter, r *http.Request) {
	parsedLayout, err := template.New("app").Funcs(template.FuncMap{
		"T": func(msg string) string {
			return WebT(r, msg)
		},
	}).Parse(p.Layout)
	if err != nil {
		http.Error(w, "Internal error (parse)", 500)
		log.Println("ERROR: parse template:", err)
	}

	if err := parsedLayout.Execute(w, p); err != nil {
		http.Error(w, "Internal error (execute)", 500)
		log.Println("ERROR: execute template:", err)
	}
}
