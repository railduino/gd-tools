package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
)

const (
	OCISName   = "ocis.json"
	OCISNumber = 25
)

type OCIS struct {
	HostName   string `json:"host_name"`
	DomainName string `json:"domain_name"`
	AdminName  string `json:"admin_name"`
	AdminEmail string `json:"admin_email"`
	Password   string `json:"password"`
	Language   string `json:"language"`
	LogLevel   string `json:"log_level"`
}

func (oc *OCIS) FQDN() string {
	return oc.HostName + "." + oc.DomainName
}

func (oc *OCIS) ExecPath() string {
	return agent.GetBinDir("ocis")
}

func (oc *OCIS) RootDir(paths ...string) string {
	rootDir := agent.GetToolsDir("data", "ocis")
	if len(paths) == 0 {
		return rootDir
	}
	return filepath.Join(append([]string{rootDir}, paths...)...)
}

func (oc *OCIS) BaseDir(paths ...string) string {
	baseDir := oc.RootDir(".ocis")
	if len(paths) == 0 {
		return baseDir
	}
	return filepath.Join(append([]string{baseDir}, paths...)...)
}

func (oc *OCIS) ConfigDir() string {
	return oc.BaseDir("config")
}

func (oc *OCIS) ClientDir(paths ...string) string {
	clientDir := filepath.Join(oc.RootDir(), "client")
	if len(paths) == 0 {
		return clientDir
	}
	return filepath.Join(append([]string{clientDir}, paths...)...)
}

func (oc *OCIS) CertDir() string {
	return agent.GetToolsDir("data", "certs", oc.FQDN())
}

func (oc *OCIS) LogsDir() string {
	return agent.GetToolsDir("logs", "ocis")
}

func (oc *OCIS) Save() error {
	content, err := json.MarshalIndent(oc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", OCISName, err)
	}

	existing, err := os.ReadFile(OCISName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(OCISName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", OCISName, err)
	}

	return nil
}

func (cfg *Config) DeployOCIS() error {
	cfg.Debug("Enter pkg/config/ocis.go")

	content, err := os.ReadFile(OCISName)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", OCISName, err)
	}

	var oc OCIS
	if err := json.Unmarshal(content, &oc); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", OCISName, err)
	}
	cfg.OCIS = &oc

	if err := cfg.OCISDownload(); err != nil {
		return err
	}

	if err := cfg.OCISUser(); err != nil {
		return err
	}

	if err := cfg.EnsureIDMSelfSignedCert(); err != nil {
		return err
	}

	if err := cfg.OCISConfig(); err != nil {
		return err
	}

	if err := cfg.OCISService(); err != nil {
		return err
	}

	if err := cfg.EnsureCertificate(oc.FQDN()); err != nil {
		return err
	}

	if err := cfg.OCISVhost(); err != nil {
		return err
	}

	if status, err := cfg.SetCNAME(oc.DomainName, oc.HostName); err != nil {
		return err
	} else if status != "" {
		cfg.Say(status)
	}

	cfg.Debug("Leave pkg/config/ocis.go")
	return nil
}

func (cfg *Config) OCISDownload() error {
	req := cfg.NewRequest()

	ocisBin, err := agent.GetDownload("ocis")
	if err != nil {
		return err
	}
	req.Downloads = append(req.Downloads, ocisBin)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) OCISUser() error {
	req := cfg.NewRequest()
	oc := cfg.OCIS

	ocisUser := agent.User{
		Name:    "ocis",
		Comment: "OCIS Server User",
		System:  true,
		HomeDir: oc.RootDir(),
		Shell:   "/bin/bash",
	}
	req.Users = append(req.Users, &ocisUser)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) OCISConfig() error {
	req := cfg.NewRequest()
	oc := cfg.OCIS

	rootMkdir := agent.File{
		Task:  "mkdir",
		Path:  oc.RootDir(),
		Mode:  "0750",
		User:  "ocis",
		Group: "ocis",
	}
	req.AddFile(&rootMkdir)

	baseMkdir := agent.File{
		Task:  "mkdir",
		Path:  oc.BaseDir(),
		Mode:  "0750",
		User:  "ocis",
		Group: "ocis",
	}
	req.AddFile(&baseMkdir)

	configMkdir := agent.File{
		Task:  "mkdir",
		Path:  oc.ConfigDir(),
		Mode:  "0750",
		User:  "ocis",
		Group: "ocis",
	}
	req.AddFile(&configMkdir)

	envTmpl := filepath.Join("ocis", "ocis.env")
	envData, err := templates.Parse(envTmpl, cfg.Verbose, oc)
	if err != nil {
		return err
	}
	envPath := filepath.Join(oc.ConfigDir(), "ocis.env")
	envFile := agent.File{
		Task:    "write",
		Path:    envPath,
		Content: envData,
		Mode:    "0640",
		User:    "ocis",
		Group:   "ocis",
	}
	req.AddFile(&envFile)

	idmMkdir := agent.File{
		Task:  "mkdir",
		Path:  oc.BaseDir("idm"),
		Mode:  "0750",
		User:  "ocis",
		Group: "ocis",
	}
	req.AddFile(&idmMkdir)

	crtName := "ocis-ldap.crt"
	crtData, err := os.ReadFile(crtName)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", crtName, err)
	}
	crtFile := agent.File{
		Task:    "write",
		Path:    oc.BaseDir("idm", "ldap.crt"),
		Content: crtData,
		Mode:    "0640",
		User:    "ocis",
		Group:   "ocis",
	}
	req.AddFile(&crtFile)

	keyName := "ocis-ldap.key"
	keyData, err := os.ReadFile(keyName)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", keyName, err)
	}
	keyFile := agent.File{
		Task:    "write",
		Path:    oc.BaseDir("idm", "ldap.key"),
		Content: keyData,
		Mode:    "0600",
		User:    "ocis",
		Group:   "ocis",
	}
	req.AddFile(&keyFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) OCISService() error {
	req := cfg.NewRequest()

	path := filepath.Join("ocis", "ocis.service")
	content, err := templates.Parse(path, cfg.Verbose, cfg.OCIS)
	if err != nil {
		return err
	}

	file := agent.File{
		Task:    "write",
		Path:    agent.GetEtcDir("systemd", "system", "ocis.service"),
		Content: content,
		Mode:    "0644",
		Service: "ocis",
	}
	req.AddFile(&file)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) OCISVhost() error {
	req := cfg.NewRequest()

	logsMkdir := agent.File{
		Task:  "mkdir",
		Path:  cfg.OCIS.LogsDir(),
		Mode:  "0750",
		User:  "www-data",
		Group: "www-data",
	}
	req.AddFile(&logsMkdir)

	vhostTmpl, err := templates.Parse("ocis/vhost.conf", cfg.Verbose, cfg.OCIS)
	if err != nil {
		return err
	}
	vhostName := fmt.Sprintf("%02d-%s.conf", OCISNumber, cfg.OCIS.FQDN())
	vhostPath := agent.GetApacheEtcDir("sites-available", vhostName)
	vhostFile := agent.File{
		Task:    "write",
		Path:    vhostPath,
		Content: vhostTmpl,
		Service: "apache2",
	}
	req.AddFile(&vhostFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
