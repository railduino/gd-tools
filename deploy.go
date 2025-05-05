package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// Typ 1: DeployTemplate → überträgt statische Dateien aus templates/ ohne Rendern
func DeployTemplate(dryRun bool, fileName, destPath, receiver, chmod string) error {
	content, err := TemplateLoad(fileName)
	if err != nil {
		return fmt.Errorf("TemplateLoad(%s) failed: %w", fileName, err)
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
		DryRun:   dryRun,
		Flags:    []string{"--chown=root:root", "--chmod=" + chmod},
		Local:    tmpFile.Name(),
		Receiver: receiver,
		Remote:   destPath,
	}

	return rsync.Execute()
}

// Typ 2: DeployParsedTemplate → rendert templates mit Platzhaltern und überträgt sie
func DeployParsedTemplate(dryRun bool, tmplName, destPath, receiver, chmod string, data any) error {
	rendered, err := TemplateParse(tmplName, data)
	if err != nil {
		return fmt.Errorf("TemplateParse(%s) failed: %w", tmplName, err)
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
		DryRun:   dryRun,
		Flags:    []string{"--chown=root:root", "--chmod=" + chmod},
		Local:    tmpFile.Name(),
		Receiver: receiver,
		Remote:   destPath,
	}

	return rsync.Execute()
}

// Typ 3: DeployLocal → überträgt lokale echte Dateien oder Verzeichnisse unverändert
func DeployLocal(dryRun bool, localPath, destPath, receiver, chmod string) error {
	rsync := DeployRsync{
		DryRun:   dryRun,
		Flags:    []string{"--chown=root:root", "--chmod=" + chmod},
		Local:    localPath,
		Receiver: receiver,
		Remote:   destPath,
	}
	return rsync.Execute()
}

// Utility: holt /etc/letsencrypt vom Zielsystem lokal
func deployFetchLetsEncrypt(dryRun bool, rootUser string) {
	rsyncCmd := fmt.Sprintf("rsync -avz %s:/etc/letsencrypt/ letsencrypt", rootUser)

	if err := ShellCmd(dryRun, rsyncCmd); err != nil {
		fmt.Println("Ignore error:", err)
	}
}
