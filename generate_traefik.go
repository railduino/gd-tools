package main

import (
	"fmt"
	"os"
	_ "path/filepath"

	"github.com/urfave/cli/v2"
)

type TraefikTemplateData struct {
	TraefikVersion string `json:"traefik_version"`
	EmailUser      string `json:"email_user"`
	LogLevel       string `json:"log_level"`
	StatusHost     string `json:"status_host"`
	StatusUser     string `json:"status_user"`
	StatusPswd     string `json:"status_pswd"`
	DataDir        string `json:"-"`
	LogsDir        string `json:"-"`
}

func init() {
	RegisterProjectKind(commandGenerateTraefik)
}

var commandGenerateTraefik = &cli.Command{
	Name:        "traefik",
	Usage:       T("install-traefik-usage"),
	Description: T("install-traefik-describe"),
	Flags: []cli.Flag{
		&generateDependsFlag,
	},
	Action: runGenerateTraefik,
}

func runGenerateTraefik(c *cli.Context) error {
	args := c.Args().Slice()
	if len(args) < 1 {
		msg := T("install-err-missing-prefix")
		return fmt.Errorf(msg)
	}

	if err := FileSystemCheck(); err != nil {
		return err
	}

	project := Project{
		Prefix: args[0],
		Kind:   "traefik",
		Name:   "",
	}

	if err := project.CheckConflict(true); err != nil {
		return err
	}

	fmt.Println(T("gen-create-dir"))
	if err := os.MkdirAll(project.GetName(), 0755); err != nil {
		return err
	}

	fmt.Println(T("gen-create-config"))
	project.Config = Config{
		IsEnabled: false,
		DependsOn: []string{},
		Networks:  []string{},
	}

	for _, depend := range c.StringSlice("depends") {
		project.Config.DependsOn = append(project.Config.DependsOn, depend)
	}

	// TODO make common network?

	if err := project.SaveConfig(); err != nil {
		return err
	}

	fmt.Println(T("gen-create-compose"))
	composeData := TraefikTemplateData{
		TraefikVersion: "v2.10",
		EmailUser:      "admin@example.com",
		LogLevel:       "INFO",
		StatusHost:     "status.example.com",
		StatusUser:     "Admin",
		StatusPswd:     "xyz123",
		DataDir:        project.GetVolumePath(true),
		LogsDir:        project.GetLogsPath(true),
	}
	// TODO find reasonable defaults (e.g. for password)

	var err error
	project.Compose, err = TemplateLoad("traefik", "compose.yaml", composeData)
	if err != nil {
		return err
	}
	if err := project.SaveCompose(); err != nil {
		return err
	}

	/* TODO add this to the deployment on the server
	acmePath := filepath.Join("config", "acme.json")
	if _, err := os.Stat(acmePath); os.IsNotExist(err) {
		if err := os.WriteFile(acmePath, []byte("{}"), 0600); err != nil {
			return fmt.Errorf("Fehler beim Schreiben von acme.json: %w", err)
		}
	}
	*/

	return nil
}
