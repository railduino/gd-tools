package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

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
)

func init() {
	AddSubCommand(commandServe, "any") // TODO "prod"
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

	serveConfig, err := ReadServeConfig()
	if err != nil {
		return err
	}

	LocaleInit()
	for _, line := range LocaleGetInfo() {
		log.Printf("Locale: %s", line)
	}

	if err := RunWebServer(serveConfig.Address); err != nil {
		return err
	}

	return nil
}

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
