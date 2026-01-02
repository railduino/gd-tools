package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
	"github.com/railduino/gd-tools/utils"
)

const (
	BackupName = "backup.json"
)

type Backup struct {
	BorgRepo string `json:"-"`
	DataDir  string `json:"-"`
	HookDir  string `json:"-"`

	PassPhrase  string `json:"pass_phrase"`
	RemotePath  string `json:"remote_path"`
	RemoteShell string `json:"remote_shell"`
	KeepDays    int    `json:"keep_days"`
	KeepWeeks   int    `json:"keep_weeks"`
	KeepMonths  int    `json:"keep_months"`
	CronHour    int    `json:"cron_hour"`
	CronMinute  int    `json:"cron_minute"`
}

func (bkup *Backup) CronLine() string {
	return fmt.Sprintf("%d %d * * * root test -x %s && %s\n",
		bkup.CronMinute,
		bkup.CronHour,
		agent.GetBinDir("bb.exec"),
		agent.GetBinDir("bb.exec"),
	)
}

func (cfg *Config) DeployBackup() error {
	cfg.Debug("Enter pkg/config/backup.go")

	bkup, err := ReadBackup()
	if err != nil {
		return err
	}

	req := cfg.NewRequest()

	if bkup.RemotePath != "" {
		bkup.BorgRepo = bkup.RemotePath
		privContent, pubContent, err := utils.GetRSAKeyPair(cfg.FQDN())
		if err != nil {
			return err
		}
		privFile := agent.File{
			Task:    "write",
			Path:    agent.GetRootDir(".ssh", "id_rsa"),
			Content: privContent,
			Mode:    "0600",
			User:    "root",
			Group:   "root",
		}
		req.AddFile(&privFile)
		pubFile := agent.File{
			Task:    "write",
			Path:    agent.GetRootDir(".ssh", "id_rsa.pub"),
			Content: pubContent,
			Mode:    "0600",
			User:    "root",
			Group:   "root",
		}
		req.AddFile(&pubFile)
	} else {
		bkup.BorgRepo = agent.GetToolsDir("backup")
		backupMkdir := agent.File{
			Task:  "mkdir",
			Path:  bkup.BorgRepo,
			Mode:  "0700",
			User:  "root",
			Group: "root",
		}
		req.AddFile(&backupMkdir)
	}

	bkup.DataDir = agent.GetToolsDir("data")
	bkup.HookDir = agent.GetToolsDir("data", "hooks")

	bbFiles := []string{
		"bb.check",
		"bb.delete",
		"bb.exec",
		"bb.info",
		"bb.list",
		"bb.mount",
	}

	for _, name := range bbFiles {
		tmpl := filepath.Join("backup", name)
		content, err := templates.Parse(tmpl, cfg.Verbose, bkup)
		if err != nil {
			return err
		}
		file := agent.File{
			Task:    "write",
			Path:    agent.GetBinDir(name),
			Content: content,
			Mode:    "0500",
			User:    "root",
			Group:   "root",
		}
		req.AddFile(&file)
	}

	cronFile := agent.File{
		Task:    "write",
		Path:    agent.GetEtcDir("cron.d/borg-backup"),
		Content: []byte(bkup.CronLine()),
		Mode:    "0644",
		User:    "root",
		Group:   "root",
	}
	req.AddFile(&cronFile)

	if err := req.Send(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/backup.go")
	return nil
}

func ReadBackup() (*Backup, error) {
	content, err := os.ReadFile(BackupName)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", BackupName, err)
	}

	var bkup Backup
	if err := json.Unmarshal(content, &bkup); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", BackupName, err)
	}

	return &bkup, nil
}

func (bkup *Backup) Save() error {
	content, err := json.MarshalIndent(bkup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", BackupName, err)
	}

	existing, err := os.ReadFile(BackupName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(BackupName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", BackupName, err)
	}

	return nil
}
