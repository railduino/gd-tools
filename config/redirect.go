package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
)

const (
	RedirectName = "redirect.json"
)

type Redirect struct {
	HostName    string   `json:"host_name"`
	DomainName  string   `json:"domain_name"`
	Aliases     []string `json:"aliases"`
	Destination string   `json:"destination"`
	AdminEmail  string   `json:"admin_email"`
}

type RedirectList struct {
	Entries []*Redirect `json:"entries"`
}

func (redir *Redirect) FQDN() string {
	return redir.HostName + "." + redir.DomainName
}

func (redir *Redirect) Name() string {
	return agent.MakeDBName(redir.FQDN())
}

func (redir *Redirect) IsWWW() bool {
	return redir.HostName == "www"
}

func (redir *Redirect) ServerAlias() string {
	aliases := []string{
		redir.FQDN(),
	}
	for _, domain := range redir.Aliases {
		aliases = append(aliases, "www."+domain, domain)
	}
	sort.Strings(aliases)
	return strings.Join(aliases, " ")
}

func (redir *Redirect) LogsDir() string {
	return agent.GetToolsDir("logs", "redirect", redir.Name())
}

func (redir *Redirect) CertDir() string {
	return agent.GetToolsDir("data", "certs", redir.FQDN())
}

func (redir *Redirect) VhostPath() string {
	name := fmt.Sprintf("07-%s.conf", redir.FQDN())
	return agent.GetApacheEtcDir("sites-available", name)
}

func (redir *Redirect) CertificateList() (string, []string) {
	if !redir.IsWWW() {
		return redir.FQDN(), nil
	}

	list := []string{
		redir.DomainName,
	}

	for _, domain := range redir.Aliases {
		list = append(list, "www."+domain, domain)
	}

	return redir.FQDN(), list
}

func LoadRedirectList() (*RedirectList, error) {
	var list RedirectList

	content, err := os.ReadFile(RedirectName)
	if err != nil {
		if os.IsNotExist(err) {
			return &list, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", RedirectName, err)
	}

	if err := json.Unmarshal(content, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", RedirectName, err)
	}

	sort.Slice(list.Entries, func(i, j int) bool {
		return list.Entries[i].FQDN() < list.Entries[j].FQDN()
	})

	return &list, nil
}

func (list *RedirectList) Save() error {
	sort.Slice(list.Entries, func(i, j int) bool {
		return list.Entries[i].FQDN() < list.Entries[j].FQDN()
	})

	content, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", RedirectName, err)
	}

	existing, err := os.ReadFile(RedirectName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(RedirectName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", RedirectName, err)
	}

	return nil
}

func (cfg *Config) DeployRedirect() error {
	cfg.Debug("Enter pkg/config/redirect.go")

	list, err := LoadRedirectList()
	if err != nil {
		return nil
	}

	for _, redirect := range list.Entries {
		if redirect.FQDN() == cfg.FQDN() {
			return fmt.Errorf("cannot use the server name for redirect")
		}

		fqdnCert, sanCerts := redirect.CertificateList()
		if err := cfg.EnsureCertificate(fqdnCert, sanCerts...); err != nil {
			return err
		}

		if err := cfg.RedirectLogsDir(redirect); err != nil {
			return err
		}

		if err := cfg.RedirectSetupVhost(redirect); err != nil {
			return err
		}

		if cfg.SkipDNS {
			cfg.Sayf("skipping dns-update for %s", redirect.FQDN())
			return nil
		}

		// install all www entries as CNAME records
		// The DomainList contains at least the FQDN of the server
		if status, err := cfg.SetCNAME(redirect.DomainName, redirect.HostName); err != nil {
			return err
		} else if status != "" {
			cfg.Say(status)
		}

		// for non-www servers (e.g. demo.example.com) there is nothing more to do
		if !redirect.IsWWW() {
			continue
		}

		// install the domain A and AAAA records (those cannot be CNAME)
		if cfg.IPv4Addr != "" {
			if status, err := cfg.SetA(redirect.DomainName, "@", cfg.IPv4Addr); err != nil {
				return err
			} else if status != "" {
				cfg.Say(status)
			}
		}
		if cfg.IPv6Addr != "" {
			if status, err := cfg.SetAAAA(redirect.DomainName, "@", cfg.IPv6Addr); err != nil {
				return err
			} else if status != "" {
				cfg.Say(status)
			}
		}
	}

	cfg.Debug("Leave pkg/config/redirect.go")
	return nil
}

func (cfg *Config) RedirectLogsDir(redir *Redirect) error {
	req := cfg.NewRequest()

	logsMkdir := agent.File{
		Task:  "mkdir",
		Path:  redir.LogsDir(),
		Mode:  "0755",
		User:  "www-data",
		Group: "www-data",
	}
	req.AddFile(&logsMkdir)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) RedirectSetupVhost(redir *Redirect) error {
	req := cfg.NewRequest()

	vhostTmpl := filepath.Join("redirect", "vhost.conf")
	vhostContent, err := templates.Parse(vhostTmpl, cfg.Verbose, redir)
	if err != nil {
		return err
	}

	if redir.AdminEmail == "" {
		redir.AdminEmail = "admin@" + redir.DomainName
	}

	vhostFile := agent.File{
		Task:    "write",
		Path:    redir.VhostPath(),
		Content: vhostContent,
	}
	req.AddFile(&vhostFile)
	req.AddService("apache2")

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
