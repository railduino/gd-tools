package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	ServeConfigName = "gd-tools-serve.conf"
)

func init() {
	AddSubCommand(commandServe, "prod")
}

var commandServe = &cli.Command{
	Name:        "serve",
	Usage:       T("serve-cmd-usage"),
	Description: T("serve-cmd-describe"),
	Flags:       []cli.Flag{},
	Action:      runServe,
}

func runServe(c *cli.Context) error {
	if euid := os.Geteuid(); euid == 0 {
		msg := T("serve-not-as-root")
		return fmt.Errorf(msg)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.Chdir(homeDir); err != nil {
		return err
	}

	sc, err := ReadServeConfig()
	if err != nil {
		return err
	}

	if sc.LayoutContent, err = ServeLoadPage("application.html"); err != nil {
		return err
	}
	if sc.HomeContent, err = ServeLoadPage("home.html"); err != nil {
		return err
	}
	if sc.StatusContent, err = ServeLoadPage("status.html"); err != nil {
		return err
	}

	sc.Mux = http.NewServeMux()
	sc.Mux.HandleFunc("/", sc.HomeHandler)
	sc.Mux.HandleFunc("/status", sc.BasicAuthMiddleware(sc.StatusHandler))

	LocaleInit()
	for _, line := range LocaleGetInfo() {
		log.Printf("Locale: %s", line)
	}

	return sc.RunWebServer()
}

func ReadServeConfig() (*ServeConfig, error) {
	var serveConfig ServeConfig

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
