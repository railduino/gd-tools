package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
)

func (cfg *Config) DeployRustDesk() error {
	cfg.Debug("Enter pkg/config/rustdesk.go")

	content, err := os.ReadFile(agent.RustDeskName)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", agent.RustDeskName, err)
	}

	var rd agent.RustDesk
	if err := json.Unmarshal(content, &rd); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", agent.RustDeskName, err)
	}
	cfg.RustDesk = &rd

	if err := cfg.RustDeskUser(); err != nil {
		return err
	}

	if err := cfg.RustDeskDownload(); err != nil {
		return err
	}

	if err := cfg.RustDeskExtract(); err != nil {
		return err
	}

	if err := cfg.RustDeskService(); err != nil {
		return err
	}

	if err := cfg.RustDeskFirewall(); err != nil {
		return err
	}

	if status, err := cfg.SetCNAME(rd.DomainName, rd.HostName); err != nil {
		return err
	} else if status != "" {
		cfg.Say(status)
	}

	cfg.Debug("Leave pkg/config/rustdesk.go")
	return nil
}

func (cfg *Config) RustDeskUser() error {
	req := cfg.NewRequest()
	rd := cfg.RustDesk

	rustdeskUser := agent.User{
		Name:    "rustdesk",
		Comment: "RustDesk Server User",
		System:  true,
		HomeDir: rd.DataDir(),
		Shell:   "/usr/sbin/nologin",
	}
	req.Users = append(req.Users, &rustdeskUser)

	dataMkdir := agent.File{
		Task:  "mkdir",
		Path:  rd.DataDir(),
		Mode:  "0750",
		User:  "rustdesk",
		Group: "rustdesk",
	}
	req.AddFile(&dataMkdir)

	logsMkdir := agent.File{
		Task:  "mkdir",
		Path:  rd.LogsDir(),
		Mode:  "0750",
		User:  "rustdesk",
		Group: "rustdesk",
	}
	req.AddFile(&logsMkdir)

	if err := req.Send(); err != nil {
		return err
	}
	return nil
}

func (cfg *Config) RustDeskDownload() error {
	req := cfg.NewRequest()

	zip, err := agent.GetDownload("rustdesk")
	if err != nil {
		return err
	}
	req.Downloads = append(req.Downloads, zip)

	if err := req.Send(); err != nil {
		return err
	}
	return nil
}

func (cfg *Config) RustDeskExtract() error {
	req := cfg.NewRequest()
	rd := cfg.RustDesk

	download, err := agent.GetDownload("rustdesk")
	if err != nil {
		return err
	}

	extract := agent.File{
		Task:   "extract",
		Path:   agent.GetDownloadsDir(download.FileName),
		Target: rd.DataDir(),
		Mode:   "0750",
		User:   "rustdesk",
		Group:  "rustdesk",
	}
	req.AddFile(&extract)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) RustDeskService() error {
	req := cfg.NewRequest()

	// hbbs unit
	{
		path := filepath.Join("rustdesk", "hbbs.service")
		content, err := templates.Parse(path, cfg.Verbose, cfg.RustDesk)
		if err != nil {
			return err
		}
		file := agent.File{
			Task:    "write",
			Path:    agent.GetEtcDir("systemd", "system", "rustdesk-hbbs.service"),
			Content: content,
			Mode:    "0644",
			Service: "rustdesk-hbbs",
		}
		req.AddFile(&file)
	}

	// hbbr unit
	{
		path := filepath.Join("rustdesk", "hbbr.service")
		content, err := templates.Parse(path, cfg.Verbose, cfg.RustDesk)
		if err != nil {
			return err
		}
		file := agent.File{
			Task:    "write",
			Path:    agent.GetEtcDir("systemd", "system", "rustdesk-hbbr.service"),
			Content: content,
			Mode:    "0644",
			Service: "rustdesk-hbbr",
		}
		req.AddFile(&file)
	}

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) RustDeskFirewall() error {
	cfg.AddFirewall("21115/tcp")
	cfg.AddFirewall("21116/tcp")
	cfg.AddFirewall("21116/udp")
	cfg.AddFirewall("21117/tcp")
	if err := cfg.Save(); err != nil {
		return err
	}

	req := cfg.NewRequest()
	req.RustDesk = cfg.RustDesk
	req.AddFirewall("21115/tcp")
	req.AddFirewall("21116/tcp")
	req.AddFirewall("21116/udp")
	req.AddFirewall("21117/tcp")
	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
