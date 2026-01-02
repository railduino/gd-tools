package config

import (
	"fmt"
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/php"
	"github.com/railduino/gd-tools/templates"
	"github.com/railduino/gd-tools/utils"
)

type Roundcube struct {
	DomainName string
	Name       string
	FQDN       string
	SysAdmin   string
	Locale     string
	Password   string
	DesKey     string
	DirName    string
}

func (rc *Roundcube) WebMail() string {
	return "webmail." + rc.DomainName
}

func (rc *Roundcube) RootDir() string {
	return agent.GetToolsDir("data", "roundcube", rc.Name)
}

func (rc *Roundcube) BaseDir() string {
	return filepath.Join(rc.RootDir(), rc.DirName)
}

func (rc *Roundcube) CertDir() string {
	return agent.GetToolsDir("data", "certs", rc.WebMail())
}

func (rc *Roundcube) LogsDir() string {
	return agent.GetToolsDir("logs", "roundcube", rc.Name)
}

func (rc *Roundcube) SocketPath() string {
	return fmt.Sprintf("/run/php/php%s-roundcube-%s.sock", php.GetPhpVersion(), rc.Name)
}

func (cfg *Config) DeployRoundcube() error {
	domainList, _, err := email.GetDomains(nil)
	if err != nil {
		return err
	}

	for _, domain := range domainList.Domains {
		if err := cfg.DeployRoundcubeDomain(domain); err != nil {
			return err
		}
	}

	return nil
}

func (cfg *Config) DeployRoundcubeMap(sel map[string]bool) error {
	domainList, _, err := email.GetDomains(sel)
	if err != nil {
		return err
	}

	for _, domain := range domainList.Domains {
		if err := cfg.DeployRoundcubeDomain(domain); err != nil {
			return err
		}
	}

	return nil
}

func (cfg *Config) DeployRoundcubeDomain(domain *email.Domain) error {
	cfg.Debugf("Enter pkg/config/roundcube.go (%s)", domain.Name)

	rc := &Roundcube{
		DomainName: domain.Name,
		Name:       agent.MakeDBName(domain.Name),
		FQDN:       cfg.FQDN(), // for Postfix / Dovecot access
		SysAdmin:   cfg.SysAdmin,
		Locale:     cfg.Locale(),
	}

	download, err := agent.GetDownload("roundcube")
	if err != nil {
		return err
	}
	if download.DirName == "" {
		return fmt.Errorf("missing DirName in Roundcube download")
	}
	rc.DirName = download.DirName

	rc.Password, err = utils.FetchPassword(20, "vmail", "db_password")
	if err != nil {
		return err
	}

	rc.DesKey, err = utils.FetchPassword(24, "roundcube-"+rc.Name, "des_key")
	if err != nil {
		return err
	}

	if err := cfg.RoundcubeDownload(); err != nil {
		return err
	}

	if err := cfg.RoundcubeExtract(rc); err != nil {
		return err
	}

	// TODO (later) replace logo, favicon in skins/elastic/images

	if err := cfg.RoundcubeDataDirs(rc); err != nil {
		return err
	}

	if err := cfg.RoundcubeConfigFiles(rc); err != nil {
		return err
	}

	if err := cfg.RoundcubeSQL(rc); err != nil {
		return err
	}

	if err := cfg.RoundcubeHideInstaller(rc); err != nil {
		return err
	}

	if err := cfg.EnsureCertificate(rc.WebMail()); err != nil {
		return err
	}

	if err := cfg.RoundcubeSetupVhost(rc); err != nil {
		return err
	}

	if status, err := cfg.SetCNAME(domain.Name, "webmail"); err != nil {
		return err
	} else if status != "" {
		cfg.Say(status)
	}

	cfg.Debugf("Leave pkg/config/roundcube.go (%s)", domain.Name)
	return nil
}

func (cfg *Config) RoundcubeDownload() error {
	req := cfg.NewRequest()

	download, err := agent.GetDownload("roundcube")
	if err != nil {
		return err
	}

	req.Downloads = append(req.Downloads, download)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) RoundcubeExtract(rc *Roundcube) error {
	req := cfg.NewRequest()

	download, err := agent.GetDownload("roundcube")
	if err != nil {
		return err
	}

	extract := agent.File{
		Task:   "extract",
		Path:   agent.GetDownloadsDir(download.FileName),
		Target: rc.RootDir(),
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

func (cfg *Config) RoundcubeDataDirs(rc *Roundcube) error {
	req := cfg.NewRequest()

	dataDirs := []string{
		rc.LogsDir(),
		filepath.Join(rc.BaseDir(), "temp"),
		filepath.Join(rc.BaseDir(), "upload"),
	}
	for _, dir := range dataDirs {
		subDir := agent.File{
			Task:  "mkdir",
			Path:  dir,
			Mode:  "0750",
			User:  "www-data",
			Group: "www-data",
		}
		req.AddFile(&subDir)
	}

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) RoundcubeSQL(rc *Roundcube) error {
	req := cfg.NewRequest()

	cfgTmpl := filepath.Join("roundcube", "create_config.sql")
	stmts, err := templates.SQL(cfgTmpl, cfg.Verbose, rc)
	if err != nil {
		return err
	}
	entry := agent.MySQL{
		Stmts:   stmts,
		Comment: "create roundcube (vmail) tables",
	}
	loader := agent.MySQL{
		DbName:  "rc_" + rc.Name,
		DbPath:  filepath.Join(rc.BaseDir(), "SQL", "mysql.initial.sql"),
		Comment: "mysql.initial.sql",
	}
	req.MySQLs = append(req.MySQLs, &entry, &loader)

	req.AddService("apache2")
	req.AddService("dovecot")

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) RoundcubeConfigFiles(rc *Roundcube) error {
	req := cfg.NewRequest()

	configTmpl, err := templates.Parse("roundcube/config.inc.php", cfg.Verbose, rc)
	if err != nil {
		return err
	}
	configFile := agent.File{
		Task:    "write",
		Path:    filepath.Join(rc.BaseDir(), "config", "config.inc.php"),
		Content: configTmpl,
		Mode:    "0644",
	}
	req.AddFile(&configFile)

	passwordTmpl, err := templates.Parse("roundcube/password.inc.php", cfg.Verbose, rc)
	if err != nil {
		return err
	}
	passwordFile := agent.File{
		Task:    "write",
		Path:    filepath.Join(rc.BaseDir(), "config", "password.inc.php"),
		Content: passwordTmpl,
		Mode:    "0644",
	}
	req.AddFile(&passwordFile)

	poolTmpl, err := templates.Parse("roundcube/php-fpm-pool.conf", cfg.Verbose, rc)
	if err != nil {
		return err
	}
	poolPath := php.GetPhpFpmPoolPath(15, "roundcube-"+rc.Name)
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

func (cfg *Config) RoundcubeHideInstaller(rc *Roundcube) error {
	req := cfg.NewRequest()

	installerTask := agent.File{
		Task:  "process",
		Path:  filepath.Join(rc.BaseDir(), "installer"),
		Mode:  "0700",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&installerTask)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) RoundcubeSetupVhost(rc *Roundcube) error {
	req := cfg.NewRequest()

	vhostTmpl, err := templates.Parse("roundcube/vhost.conf", cfg.Verbose, rc)
	if err != nil {
		return err
	}

	vhostName := fmt.Sprintf("15-%s.conf", rc.WebMail())
	vhostPath := agent.GetApacheEtcDir("sites-available", vhostName)
	vhostFile := agent.File{
		Task:    "write",
		Path:    vhostPath,
		Content: vhostTmpl,
		Service: "apache2",
	}
	req.AddFile(&vhostFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
