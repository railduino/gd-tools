package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/php"
	"github.com/railduino/gd-tools/templates"
)

func (cfg *Config) DeployNextcloud() error {
	nc := cfg.Nextcloud
	if nc == nil {
		return fmt.Errorf("missing Nextcloud pointer")
	}
	cfg.Debugf("Enter pkg/config/nextcloud.go (%s)", nc.FQDN())

	if nc.FQDN() == cfg.FQDN() {
		return fmt.Errorf("cannot use the server name for Nextcloud")
	}

	if err := cfg.NextcloudDownload(); err != nil {
		return err
	}
	if err := cfg.NextcloudExtract(nc); err != nil {
		return err
	}

	if err := cfg.NextcloudLogsDir(nc); err != nil {
		return err
	}

	if err := cfg.NextcloudSQL(nc); err != nil {
		return err
	}

	if err := cfg.NextcloudBackupHook(nc); err != nil {
		return err
	}

	if err := cfg.NextcloudSetupConfig(nc, "maintenance:install"); err != nil {
		return err
	}

	for _, entry := range nc.GetConfigList() {
		if err := cfg.NextcloudSetupConfig(nc, entry.Key); err != nil {
			return err
		}
	}

	if err := cfg.NextcloudCronJob(nc); err != nil {
		return err
	}

	if err := cfg.EnsureCertificate(nc.FQDN()); err != nil {
		return err
	}

	if err := cfg.NextcloudSetupPool(nc); err != nil {
		return err
	}

	if err := cfg.NextcloudVhost(nc); err != nil {
		return err
	}

	if status, err := cfg.SetCNAME(nc.DomainName, nc.HostName); err != nil {
		return err
	} else if status != "" {
		cfg.Say(status)
	}

	// if all went well, install the occ-<name> command
	occSrc := agent.GetBinDir("gd-occ")
	occDst := agent.GetBinDir("occ-" + nc.Name)
	if _, err := cfg.LocalCommand(
		"rsync",
		cfg.RsyncFlags(),
		"--chown=root:root",
		"--chmod=0500",
		occSrc,
		cfg.RootUser()+":"+occDst,
	); err != nil {
		return fmt.Errorf("failed to install %s: %w", occDst, err)
	}

	cfg.Debugf("Leave pkg/config/nextcloud.go (%s)", nc.FQDN())
	return nil
}

func (cfg *Config) NextcloudDownload() error {
	req := cfg.NewRequest()

	download, err := agent.GetDownload("nextcloud")
	if err != nil {
		return err
	}
	if download.DirName == "" {
		return fmt.Errorf("missing DirName in Nextcloud download")
	}

	req.Downloads = append(req.Downloads, download)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) NextcloudExtract(nc *agent.Nextcloud) error {
	req := cfg.NewRequest()

	download, err := agent.GetDownload("nextcloud")
	if err != nil {
		return err
	}

	extract := agent.File{
		Task:   "extract",
		Path:   agent.GetDownloadsDir(download.FileName),
		Target: nc.RootDir(),
		Mode:   "0755",
		User:   "root",
		Group:  "root",
	}
	req.AddFile(&extract)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) NextcloudLogsDir(nc *agent.Nextcloud) error {
	req := cfg.NewRequest()

	logsMkdir := agent.File{
		Task:  "mkdir",
		Path:  nc.LogsDir(),
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&logsMkdir)

	dataMkdir := agent.File{
		Task:  "mkdir",
		Path:  nc.DataDir(""),
		Mode:  "0700",
		User:  "www-data",
		Group: "www-data",
	}
	req.AddFile(&dataMkdir)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) NextcloudSQL(nc *agent.Nextcloud) error {
	req := cfg.NewRequest()

	sqlTmpl := filepath.Join("nextcloud", "create.sql")
	sqlStmts, err := templates.SQL(sqlTmpl, cfg.Verbose, nc)
	if err != nil {
		return err
	}

	sql := agent.MySQL{
		Stmts:   sqlStmts,
		Comment: fmt.Sprintf("create nextcloud (%s) tables", cfg.Nextcloud.Name),
	}
	req.MySQLs = append(req.MySQLs, &sql)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) NextcloudBackupHook(nc *agent.Nextcloud) error {
	req := cfg.NewRequest()

	hookTmpl, err := templates.Parse("nextcloud/backup", cfg.Verbose, nc)
	if err != nil {
		return err
	}

	hookFile := agent.File{
		Task:    "write",
		Path:    nc.HookPath(),
		Mode:    "0500",
		Content: hookTmpl,
		User:    "root",
		Group:   "root",
	}
	req.AddFile(&hookFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) NextcloudSetupConfig(nc *agent.Nextcloud, key string) error {
	req := cfg.NewRequest()

	req.Nextcloud = nc
	req.NextConf = key

	if key == "maintenance:install" {
		content, err := json.MarshalIndent(nc, "", "  ")
		if err != nil {
			return err
		}
		configFile := agent.File{
			Task:    "write",
			Path:    nc.ConfigPath(),
			Content: content,
			Mode:    "0600",
			User:    "root",
			Group:   "root",
		}
		req.AddFile(&configFile)
	}

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) NextcloudCronJob(nc *agent.Nextcloud) error {
	req := cfg.NewRequest()

	cronTmpl, err := templates.Parse("nextcloud/cron.d", cfg.Verbose, nc)
	if err != nil {
		return err
	}

	cronFile := agent.File{
		Task:    "write",
		Path:    nc.CronPath(),
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

func (cfg *Config) NextcloudSetupPool(nc *agent.Nextcloud) error {
	req := cfg.NewRequest()

	poolTmpl, err := templates.Parse("nextcloud/php-fpm-pool.conf", cfg.Verbose, nc)
	if err != nil {
		return err
	}

	poolPath := php.GetPhpFpmPoolPath(nc.Number, nc.Name)
	poolFile := agent.File{
		Task:    "write",
		Path:    poolPath,
		Content: poolTmpl,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
		Service: php.GetPhpFpmService(),
	}
	req.AddFile(&poolFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) NextcloudVhost(nc *agent.Nextcloud) error {
	req := cfg.NewRequest()

	vhostTmpl, err := templates.Parse("nextcloud/vhost.conf", cfg.Verbose, nc)
	if err != nil {
		return err
	}

	vhostFile := agent.File{
		Task:    "write",
		Path:    nc.VhostPath(),
		Content: vhostTmpl,
	}
	req.AddFile(&vhostFile)
	req.AddService("apache2")

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
