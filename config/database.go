package config

import (
	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
)

func (cfg *Config) DeployDatabase() error {
	cfg.Debug("Enter pkg/config/database.go")

	req := cfg.NewRequest()
	req.RedisPort = 6379

	path := "mysql/mariadb.conf.d/90-server-port.cnf"
	content, err := templates.Load(path, cfg.Verbose)
	if err != nil {
		return err
	}
	file := agent.File{
		Task:    "write",
		Path:    agent.GetEtcDir(path),
		Content: content,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
		Service: "mariadb",
	}
	req.AddFile(&file)

	if err := req.Send(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/database.go")
	return nil
}
