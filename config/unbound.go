package config

import (
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
)

func (cfg *Config) DeployUnbound() error {
	cfg.Debug("Enter pkg/config/unbound.go")

	req := cfg.NewRequest()

	hintsTmpl := filepath.Join("unbound", "root.hints")
	hintsData, err := templates.Load(hintsTmpl, cfg.Verbose)
	if err != nil {
		return err
	}
	hintsPath := agent.GetVarDir("lib", "unbound", "root.hints")
	hintsFile := agent.File{
		Task:    "write",
		Path:    hintsPath,
		Content: hintsData,
		Mode:    "0644",
		User:    "unbound",
		Group:   "unbound",
	}
	req.AddFile(&hintsFile)

	serverTmpl := filepath.Join("unbound", "server.conf")
	serverData, err := templates.Load(serverTmpl, cfg.Verbose)
	if err != nil {
		return err
	}
	serverPath := agent.GetEtcDir("unbound", "unbound.conf.d", "00-server.conf")
	serverFile := agent.File{
		Task:    "write",
		Path:    serverPath,
		Content: serverData,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
		Service: "unbound",
	}
	req.AddFile(&serverFile)

	resolvDir := agent.GetEtcDir("systemd", "resolved.conf.d")
	resolvMkdir := agent.File{
		Task:  "mkdir",
		Path:  resolvDir,
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&resolvMkdir)

	resolvTmpl := filepath.Join("unbound", "resolv.conf")
	resolvData, err := templates.Load(resolvTmpl, cfg.Verbose)
	if err != nil {
		return err
	}
	resolvPath := filepath.Join(resolvDir, "99-unbound.conf")
	resolvFile := agent.File{
		Task:    "write",
		Path:    resolvPath,
		Content: resolvData,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
		Service: "systemd-resolved",
	}
	req.AddFile(&resolvFile)

	if err := req.Send(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/unbound.go")
	return nil
}
