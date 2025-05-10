package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

// DeployRsync beschreibt einen generischen rsync-Vorgang
type DeployRsync struct {
	DryRun   bool
	Flags    []string
	Local    string
	Receiver string
	Remote   string
}

// Execute führt den rsync-Befehl aus
func (rs *DeployRsync) Execute() error {
	flags := strings.Join(rs.Flags, " ")
	target := fmt.Sprintf("%s:%s", rs.Receiver, rs.Remote)
	cmd := fmt.Sprintf("rsync -avz %s %s %s", flags, rs.Local, target)
	return ShellCmd(rs.DryRun, cmd)
}

// Typ 1: DeployTemplate: überträgt statische Dateien aus templates/ ohne Rendern
func DeployTemplate(c *cli.Context, fileName, destPath, receiver, chmod string) error {
	content, err := TemplateLoad(fileName)
	if err != nil {
		return fmt.Errorf("TemplateLoad(%s) failed: %s", fileName, err.Error())
	}

	tmpFile, err := os.CreateTemp("", filepath.Base(fileName)+"-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	rsync := DeployRsync{
		DryRun: c.Bool("dry"),
		Flags: []string{
			"--chown=root:root",
			"--chmod=" + chmod,
		},
		Local:    tmpFile.Name(),
		Receiver: receiver,
		Remote:   destPath,
	}
	if !c.Bool("debug") {
		rsync.Flags = append(rsync.Flags, "--quiet")
	}

	return rsync.Execute()
}

// Typ 2: DeployParsedTemplate: rendert templates mit Platzhaltern und überträgt sie
func DeployParsedTemplate(c *cli.Context, tmplName, destPath, receiver, chmod string, data any) error {
	rendered, err := TemplateParse(tmplName, data)
	if err != nil {
		return fmt.Errorf("TemplateParse(%s) failed: %s", tmplName, err.Error())
	}

	tmpFile, err := os.CreateTemp("", filepath.Base(tmplName)+"-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(rendered); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	rsync := DeployRsync{
		DryRun: c.Bool("dry"),
		Flags: []string{
			"--chown=root:root",
			"--chmod=" + chmod,
		},
		Local:    tmpFile.Name(),
		Receiver: receiver,
		Remote:   destPath,
	}
	if !c.Bool("debug") {
		rsync.Flags = append(rsync.Flags, "--quiet")
	}

	return rsync.Execute()
}

// Typ 3: DeployLocal: überträgt lokale echte Dateien oder Verzeichnisse unverändert
func DeployLocal(c *cli.Context, localPath, destPath, receiver, chmod string) error {
	rsync := DeployRsync{
		DryRun: c.Bool("dry"),
		Flags: []string{
			"--chown=root:root",
			"--chmod=" + chmod,
		},
		Local:    localPath,
		Receiver: receiver,
		Remote:   destPath,
	}
	if !c.Bool("debug") {
		rsync.Flags = append(rsync.Flags, "--quiet")
	}

	return rsync.Execute()
}

// Utility: holt /etc/letsencrypt vom Zielsystem nach lokal
func DeployFetchLetsEncrypt(c *cli.Context, rootUser string) {
	dryRun := c.Bool("dry")

	rsyncPrefix := "rsync -avz"
	if !c.Bool("debug") {
		rsyncPrefix += " --quiet"
	}
	rsyncCmd := fmt.Sprintf("%s %s:/etc/letsencrypt/ letsencrypt", rsyncPrefix, rootUser)

	if err := ShellCmd(dryRun, rsyncCmd); err != nil {
		fmt.Println("Ignore error:", err)
	}

	if !dryRun {
		if systemConfig, err := ReadSystemConfig(false); err == nil {
			systemConfig.Save()
		}
	}
}
