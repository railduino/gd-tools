package config

import (
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/templates"
	"github.com/railduino/gd-tools/utils"
)

func DKIM_Base() string {
	return agent.GetVarDir("lib", "rspamd", "dkim")
}

func (cfg *Config) DeployRspamd() error {
	cfg.Debug("Enter pkg/config/rspamd.go")

	req := cfg.NewRequest()

	secrets, err := utils.LoadSecrets()
	if err != nil {
		return err
	}

	domainList, _, err := email.GetDomains(nil)
	if err != nil {
		return err
	}
	cfg.DKIMs = domainList.GetDKIMs(DKIM_Base())

	if entry := secrets.Get("rspamd", "password"); entry != nil {
		cfg.Password = entry.Output
	} else {
		_, cfg.Password, err = secrets.SetRspamdPassword("")
		if err != nil {
			return err
		}
	}

	dkimMkdir := agent.File{
		Task:  "mkdir",
		Path:  DKIM_Base(),
		Mode:  "0750",
		User:  "_rspamd",
		Group: "_rspamd",
	}
	req.AddFile(&dkimMkdir)

	for _, domain := range domainList.Domains {
		dkimPath := filepath.Join(DKIM_Base(), domain.NameDKIM())
		dkimFile := agent.File{
			Task:    "write",
			Path:    dkimPath,
			Content: []byte(domain.DKIM.PrivKey),
			Mode:    "0600",
			User:    "_rspamd",
			Group:   "_rspamd",
		}
		req.AddFile(&dkimFile)
	}

	files := []string{
		"actions.conf",
		"arc.conf",
		"classifier-bayes.conf",
		"dkim_signing.conf",
		"fuzzy_check.conf",
		"milter_headers.conf",
		"options.inc",
	}
	etcRspamd := agent.GetEtcDir("rspamd/local.d")

	for _, name := range files {
		path := filepath.Join("rspamd", name)
		content, err := templates.Parse(path, cfg.Verbose, cfg)
		if err != nil {
			return err
		}

		file := agent.File{
			Task:    "write",
			Path:    filepath.Join(etcRspamd, name),
			Content: content,
			Mode:    "0644",
			Service: "rspamd",
		}
		req.AddFile(&file)
	}

	if err := req.Send(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/rspamd.go")
	return nil
}
