package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

var traefikFlagHost = cli.StringFlag{
	Name:  "host",
	Usage: T("traefik-flag-status-host"),
}

var traefikFlagUser = cli.StringFlag{
	Name:  "user",
	Usage: T("traefik-flag-status-user"),
}

var traefikFlagPswd = cli.StringFlag{
	Name:  "pswd",
	Usage: T("traefik-flag-status-pswd"),
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
		&traefikFlagHost,
		&traefikFlagUser,
		&traefikFlagPswd,
	},
	Action: runGenerateTraefik,
}

func runGenerateTraefik(c *cli.Context) error {
	args := c.Args().Slice()
	if len(args) < 1 {
		msg := T("install-err-missing-prefix")
		return fmt.Errorf(msg)
	}

	systemConfig, err := FileSystemRead()
	if err != nil {
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
	statusHost := fmt.Sprintf("status.%s", systemConfig.DomainName)
	if argHost := c.String("host"); argHost != "" {
		statusHost = argHost
	}

	statusUser := fmt.Sprintf("admin@%s", systemConfig.DomainName)
	if argUser := c.String("host"); argUser != "" {
		statusUser = argUser
	}

	statusPswd := c.String("pswd")
	var secret *Secret
	if statusPswd != "" {
		secret = &Secret{
			Domain: statusHost,
			User:   statusUser,
			Input:  statusPswd,
		}
		hash, err := GetSecret(statusPswd, "bcrypt")
		if err != nil {
			return err
		}
		if err := SaveSecrets(project.GetName(), []Secret{*secret}); err != nil {
			return err
		}
		secret.Output = hash
	} else {
		secret, err = LoadSecret(project.GetName(), statusHost, statusUser)
		if err != nil {
			pw := GenerateRandomPassword()
			hash, err := GetSecret(pw, "bcrypt")
			if err != nil {
				return err
			}
			secret = &Secret{
				Domain: statusHost,
				User:   statusUser,
				Input:  pw,
				Output: hash,
			}
			if err := SaveSecrets(project.GetName(), []Secret{*secret}); err != nil {
				return err
			}
			statusPswd = pw
		} else {
			statusPswd = secret.Input
		}
	}

	dataDir, err := project.GetDataPath("prod")
	if err != nil {
		return err
	}
	logsDir, err := project.GetLogsPath("prod")
	if err != nil {
		return err
	}

	composeData := TraefikTemplateData{
		TraefikVersion: "v2.11",
		EmailUser:      systemConfig.SysAdmin,
		LogLevel:       "INFO",
		StatusHost:     statusHost,
		StatusUser:     statusUser,
		StatusPswd:     strings.ReplaceAll(secret.Output, "$", "$$"),
		DataDir:        dataDir,
		LogsDir:        logsDir,
	}

	composePath := filepath.Join("traefik", "compose.yaml")
	project.Compose, err = TemplateParse(composePath, composeData)
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
