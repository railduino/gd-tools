package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

const (
	ServeConfigName = "gd-tools-serve.json"
)

var serveConfig ServeConfig

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

	content, err := os.ReadFile(filepath.Join("/etc", ServeConfigName))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, &serveConfig); err != nil {
		return err
	}

	if serveConfig.SysAdmin == "" {
		return fmt.Errorf("missing SysAdmin")
	}
	if serveConfig.Address == "" {
		serveConfig.Address = "127.0.0.1:3000"
	}

	if serveLayoutContent, err = ServeLoadPage("application.html"); err != nil {
		return err
	}
	if serveHomeContent, err = ServeLoadPage("home.html"); err != nil {
		return err
	}
	if serveStatusContent, err = ServeLoadPage("status.html"); err != nil {
		return err
	}

	LocaleInit()
	for _, line := range LocaleGetInfo() {
		log.Printf("Locale: %s", line)
	}

	return RunWebServer()
}

func (sc ServeConfig) Save() error {
	content, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(ServeConfigName, content, 0644); err != nil {
		return err
	}

	return nil
}
