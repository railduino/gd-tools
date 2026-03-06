package config

import (
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
)

func (cfg *Config) DeployPackages(upgrade bool) error {
	cfg.Debug("Enter pkg/config/packages.go")

	if err := cfg.PackagesRepos(); err != nil {
		return err
	}

	task := "installing"
	if cfg.CheckRemote("test -r /root/.gd-tools-first-run") {
		task = "checking"
	}
	cfg.Sayf("%s %d packages - please be patient ...", task, len(cfg.Packages))

	req := cfg.NewRequest()
	req.Packages = cfg.Packages
	req.Upgrade = upgrade
	req.Firewall = cfg.Firewall
	req.UbuntuPro = cfg.UbuntuPro

	if err := req.Send(); err != nil {
		return err
	}

	cfg.PushCerts()

	cfg.Debug("Leave pkg/config/packages.go")
	return nil
}

func (cfg *Config) PackagesRepos() error {
	req := cfg.NewRequest()

	repos := []string{
		"collaboraonline",
		"docker",
		"dovecot",
	}

	for _, name := range repos {
		keyName := name + ".gpg"
		keyTmpl := filepath.Join("apt", "keys", keyName)
		keyData, err := templates.Load(keyTmpl, cfg.Verbose)
		if err != nil {
			return err
		}
		req.AddFile(&agent.File{
			Task:    "write",
			Path:    agent.GetEtcDir("apt", "keyrings", keyName),
			Content: keyData,
			Mode:    "0644",
		})

		srcName := name + ".sources"
		oldName := name + ".list"
		srcTmpl := filepath.Join("apt", "sources", srcName)
		srcData, err := templates.Load(srcTmpl, cfg.Verbose)
		if err != nil {
			return err
		}
		req.AddFile(&agent.File{
			Task:    "write",
			Path:    agent.GetEtcDir("apt", "sources.list.d", srcName),
			Content: srcData,
			Mode:    "0644",
		})
		req.AddFile(&agent.File{
			Task: "delete",
			Path: agent.GetEtcDir("apt", "sources.list.d", oldName),
		})
	}

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
