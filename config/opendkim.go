package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/templates"
)

const (
	OpenDKIMPort = 8891
)

func (cfg *Config) OpenDKIMPort() int {
	return OpenDKIMPort
}

func (cfg *Config) DeployOpenDKIM() error {
	cfg.Debug("Enter config/opendkim.go")

	req := cfg.NewRequest()

	dkimDir := agent.GetToolsDir("data", "opendkim")
	dkimMkdir := agent.File{
		Task:  "mkdir",
		Path:  dkimDir,
		Mode:  "0750",
		User:  "root",
		Group: "opendkim",
	}
	req.AddFile(&dkimMkdir)

	keysDir := filepath.Join(dkimDir, "keys")
	keysMkdir := agent.File{
		Task:  "mkdir",
		Path:  keysDir,
		Mode:  "0750",
		User:  "root",
		Group: "opendkim",
	}
	req.AddFile(&keysMkdir)

	var keyTable []string
	var signingTable []string

	domains, err := cfg.Domains()
	if err != nil {
		return err
	}
	for _, dom := range domains {
		dkim, err := dom.EnsureLocalDKIM(false)
		if err != nil {
			return fmt.Errorf("opendkim: failed to ensure DKIM for %s: %w", dom.Name, err)
		}
		if dkim == nil || dkim.PrivKey == "" {
			return fmt.Errorf("opendkim: missing private key for %s", dom.Name)
		}

		domDir := filepath.Join(keysDir, dom.Name)
		domMkdir := agent.File{
			Task:  "mkdir",
			Path:  domDir,
			Mode:  "0750",
			User:  "root",
			Group: "opendkim",
		}
		req.AddFile(&domMkdir)

		domPath := filepath.Join(domDir, email.DKIM_Selector+".private")
		domFile := agent.File{
			Task:    "write",
			Path:    domPath,
			Content: []byte(dkim.PrivKey),
			Mode:    "0640",
			User:    "root",
			Group:   "opendkim",
			Service: "opendkim",
		}
		req.AddFile(&domFile)

		keyTableLine := fmt.Sprintf("%s._domainkey.%s %s:%s:%s",
			email.DKIM_Selector, dom.Name,
			dom.Name, email.DKIM_Selector, domPath,
		)
		keyTable = append(keyTable, keyTableLine)

		signingTableLine := fmt.Sprintf("*@%s %s._domainkey.%s",
			dom.Name, email.DKIM_Selector, dom.Name,
		)
		signingTable = append(signingTable, signingTableLine)
	}

	confTmpl := filepath.Join("opendkim", "opendkim.conf")
	confData, err := templates.Parse(confTmpl, cfg.Verbose, cfg)
	if err != nil {
		return err
	}
	confFile := agent.File{
		Task:    "write",
		Path:    agent.GetEtcDir("opendkim.conf"),
		Content: confData,
		Backup:  true,
		Mode:    "0644",
		Service: "opendkim",
	}
	req.AddFile(&confFile)

	etcMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetEtcDir("opendkim"),
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&etcMkdir)

	trustedTmpl := filepath.Join("opendkim", "trusted.hosts")
	trustedData, err := templates.Parse(trustedTmpl, cfg.Verbose, cfg)
	if err != nil {
		return err
	}
	trustedFile := agent.File{
		Task:    "write",
		Path:    agent.GetEtcDir("opendkim", "TrustedHosts"),
		Content: trustedData,
		Mode:    "0644",
		Service: "opendkim",
	}
	req.AddFile(&trustedFile)

	keysData := strings.Join(keyTable, "\n") + "\n"
	keysFile := agent.File{
		Task:    "write",
		Path:    agent.GetEtcDir("opendkim", "KeyTable"),
		Content: []byte(keysData),
		Mode:    "0644",
		Service: "opendkim",
	}
	req.AddFile(&keysFile)

	signingData := strings.Join(signingTable, "\n") + "\n"
	signingFile := agent.File{
		Task:    "write",
		Path:    agent.GetEtcDir("opendkim", "SigningTable"),
		Content: []byte(signingData),
		Mode:    "0644",
		Service: "opendkim",
	}
	req.AddFile(&signingFile)

	if err := req.Send(); err != nil {
		return err
	}

	cfg.Debug("Leave config/opendkim.go")
	return nil
}
