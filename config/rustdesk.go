package config

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"os"
	// "path/filepath"

	"github.com/railduino/gd-tools/agent"
	// "github.com/railduino/gd-tools/templates"
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

	if err := cfg.RustDeskDownload(); err != nil {
		return err
	}

	if err := cfg.RustDeskUser(); err != nil {
		return err
	}

	if err := cfg.RustDeskConfig(); err != nil {
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

func (cfg *Config) RustDeskDownload() error {
	req := cfg.NewRequest()

	// downloads.json entry:
	// { "name": "rustdesk", "url": "...rustdesk-server-linux-amd64.zip", ... }
	zip, err := agent.GetDownload("rustdesk")
	if err != nil {
		return err
	}
	req.Downloads = append(req.Downloads, zip)

	// IMPORTANT:
	// This download is a ZIP containing hbbs/hbbr. After downloading,
	// the prod side must extract and install binaries into /usr/local/bin.
	//
	// If your agent already supports "unzip + install" as file tasks, add them here.
	// Otherwise implement it in the prod download handler for "rustdesk".
	//
	// Example (only if you support such tasks):
	// req.AddFile(&agent.File{Task:"unzip", Path:"/var/gd-tools/cache/rustdesk.zip", Dest:"/var/gd-tools/cache/rustdesk-unpack"})
	// req.AddFile(&agent.File{Task:"install", Path:"/var/gd-tools/cache/rustdesk-unpack/hbbs", Dest:"/usr/local/bin/hbbs", Mode:"0755"})
	// req.AddFile(&agent.File{Task:"install", Path:"/var/gd-tools/cache/rustdesk-unpack/hbbr", Dest:"/usr/local/bin/hbbr", Mode:"0755"})

	if err := req.Send(); err != nil {
		return err
	}
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

	if err := req.Send(); err != nil {
		return err
	}
	return nil
}

func (cfg *Config) RustDeskConfig() error {
	/* TODO
	req := cfg.NewRequest()

	dataMkdir := agent.File{
		Task:  "mkdir",
		Path:  rustDeskDataDir(),
		Mode:  "0750",
		User:  "rustdesk",
		Group: "rustdesk",
	}
	req.AddFile(&dataMkdir)

	logsMkdir := agent.File{
		Task:  "mkdir",
		Path:  rustDeskLogsDir(),
		Mode:  "0750",
		User:  "rustdesk",
		Group: "rustdesk",
	}
	req.AddFile(&logsMkdir)

	if err := req.Send(); err != nil {
		return err
	}
	*/
	return nil
}

func (cfg *Config) RustDeskService() error {
	/* TODO
	req := cfg.NewRequest()

	// hbbs unit
	{
		path := filepath.Join("rustdesk", "rustdesk-hbbs.service")
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
		path := filepath.Join("rustdesk", "rustdesk-hbbr.service")
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
	*/
	return nil
}

func (cfg *Config) RustDeskFirewall() error {
	/* TODO
	req := cfg.NewRequest()

	// Minimal ports for typical hbbs/hbbr usage.
	req.Firewall = append(req.Firewall,
		"21115/tcp",
		"21116/tcp",
		"21116/udp",
		"21117/tcp",
	)

	if err := req.Send(); err != nil {
		return err
	}
	*/
	return nil
}
