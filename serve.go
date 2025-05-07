package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed www/**
var wwwFS embed.FS

const (
	ServeConfigName = "gd-tools-serve.conf"
)

type ServeConfig struct {
	SysAdmin string `json:"sys_admin"`
	Password string `json:"password"`

	Address string `json:"address"`

	ImprintURL string `json:"imprint_url"`
	ProtectURL string `json:"protect_url"`
}

var (
	serveConfig ServeConfig
	serveMux    *http.ServeMux
	appTemplate string
)

func ReadServeConfig() (*ServeConfig, error) {
	content, err := os.ReadFile(ServeConfigName)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(content, &serveConfig); err != nil {
		return nil, err
	}

	return &serveConfig, nil
}

func (sc *ServeConfig) Save() error {
	content, err := json.MarshalIndent(*sc, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(ServeConfigName, content, 0644); err != nil {
		return err
	}

	return nil
}

func RunWebServer(serveAddr string) error {
	serveMux = http.NewServeMux()

	subFS, _ := fs.Sub(wwwFS, "www")
	staticServer := http.FileServer(http.FS(subFS))
	serveMux.Handle("/static/", http.StripPrefix("/", staticServer))

	appTmplFile := "www/application.html"
	appTmplContent, err := wwwFS.ReadFile(appTmplFile)
	if err != nil {
		return err
	}
	appTemplate = string(appTmplContent)

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
