package config

import (
	"fmt"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
)

func (cms *CMS) SaltEntry(num int) string {
	return fmt.Sprintf("%s_%d", cms.Salt, num)
}

func (cms *CMS) WP_CLI_Path() string {
	return agent.GetBinDir("wp-" + cms.Name())
}

func (cfg *Config) WordPressExtras(cms *CMS) error {
	wpSrc := agent.GetBinDir("gd-wp-cli")
	wpDst := agent.GetBinDir("wp-" + cms.Name())
	if _, err := cfg.LocalCommand(
		"rsync",
		cfg.RsyncFlags(),
		"--chown=root:root",
		"--chmod=0755",
		wpSrc,
		cfg.RootUser()+":"+wpDst,
	); err != nil {
		return fmt.Errorf("failed to install %s: %w", wpDst, err)
	}

	req := cfg.NewRequest()

	cronTmpl, err := templates.Parse("wordpress/cron.d", cfg.Verbose, cms)
	if err != nil {
		return err
	}
	cronFile := agent.File{
		Task:    "write",
		Path:    cms.CronPath(),
		Content: cronTmpl,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
	}
	req.AddFile(&cronFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
