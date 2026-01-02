package config

import (
	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/utils"
)

func (cfg *Config) DeployFilesystem() error {
	cfg.Debug("Enter pkg/config/filesystem.go")

	cfg.Debug("genrating DH-Parameters")
	dhBytes, err := utils.GenerateDHParams(2048)
	if err != nil {
		return err
	}

	req := cfg.NewRequest()

	vmailUser := agent.User{
		Name:    "vmail",
		Comment: "Virtual Mail User",
		System:  true,
		Shell:   "/usr/sbin/nologin",
	}
	req.Users = append(req.Users, &vmailUser)

	dataMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetToolsDir("data"),
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&dataMkdir)

	certsMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetToolsDir("data", "certs"),
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&certsMkdir)

	hooksMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetToolsDir("data", "hooks"),
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&hooksMkdir)

	vmailMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetToolsDir("data", "vmail"),
		Mode:  "0755",
		User:  "vmail",
		Group: "vmail",
	}
	req.AddFile(&vmailMkdir)

	logsMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetToolsDir("logs"),
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&logsMkdir)

	sshMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetRootDir(".ssh"),
		Mode:  "0700",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&sshMkdir)

	dhFile := agent.File{
		Task:    "write",
		Path:    cfg.DHParamsPath(),
		Content: dhBytes,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
	}
	req.AddFile(&dhFile)

	if err := req.Send(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/filesystem.go")
	return nil
}
