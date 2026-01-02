package config

import (
	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/php"
	"github.com/railduino/gd-tools/templates"
)

func (cfg *Config) DeployPHP() error {
	cfg.Debug("Enter pkg/config/php.go")

	req := cfg.NewRequest()

	content, err := templates.Parse("php/custom.ini", cfg.Verbose, cfg)
	if err != nil {
		return err
	}

	phpDirs := []string{
		"cli",
		"fpm",
	}

	for _, name := range phpDirs {
		service := "apache2"
		if name == "fpm" {
			service = php.GetPhpFpmService()
		}
		file := agent.File{
			Task:    "write",
			Path:    php.GetPhpEtcDir(name, "conf.d", "60-custom.ini"),
			Content: content,
			Mode:    "0644",
			Service: service,
		}
		req.AddFile(&file)
	}

	if err := req.Send(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/php.go")
	return nil
}
