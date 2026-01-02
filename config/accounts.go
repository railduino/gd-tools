package config

import (
	"fmt"
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/templates"
)

type Mailer struct {
	VmailUID string
	VmailGID string
	MailPath string
}

func LoadMailer() (*Mailer, error) {
	vmailUser, err := agent.GetUserID("vmail")
	if err != nil {
		return nil, err
	}

	mailer := Mailer{
		VmailUID: vmailUser.UID,
		VmailGID: vmailUser.GID,
		MailPath: agent.GetToolsDir("data", "vmail"),
	}

	return &mailer, nil
}

func (cfg *Config) DeployAccounts() error {
	return cfg.DeployAccountMap(nil)
}

func (cfg *Config) DeployAccountMap(sel map[string]bool) error {
	cfg.Debugf("Enter pkg/config/accounts.go (%v)", sel)

	mailer, err := LoadMailer()
	if err != nil {
		return err
	}
	cfg.Mailer = mailer

	domainList, _, err := email.GetDomains(sel)
	if err != nil {
		return err
	}

	for _, domain := range domainList.Domains {
		if err := cfg.DeployAccountDomain(domain); err != nil {
			return err
		}
		if err := cfg.DeployAccountProxy(domain); err != nil {
			return err
		}
	}

	emailSANs := email.GetDomainSANs()
	if err := cfg.EnsureCertificate(cfg.FQDN(), emailSANs...); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/accounts.go")
	return nil
}

func (cfg *Config) DeployAccountDomain(domain *email.Domain) error {
	req := cfg.NewRequest()

	domainDir := agent.GetToolsDir("data", "vmail", domain.Name)
	domainMkdir := agent.File{
		Task:  "mkdir",
		Path:  domainDir,
		Mode:  "0750",
		User:  "vmail",
		Group: "vmail",
	}
	req.AddFile(&domainMkdir)

	req.AddService("postfix")
	req.AddService("dovecot")

	if err := req.Send(); err != nil {
		return err
	}

	if err := cfg.ProxyDirs(domain); err != nil {
		return err
	}
	if err := cfg.ProxyCert(domain); err != nil {
		return err
	}
	if err := cfg.ProxyVhost(domain); err != nil {
		return err
	}

	for _, user := range domain.UserList {
		if err := cfg.EmailUser(user); err != nil {
			return err
		}
	}

	return nil
}

func (cfg *Config) EmailUser(user *email.User) error {
	req := cfg.NewRequest()

	userDir := agent.GetToolsDir("data", "vmail", user.Domain, user.Local)
	userSubDirs := []string{
		userDir,
		filepath.Join(userDir, "Maildir"),
		filepath.Join(userDir, "Maildir", "cur"),
		filepath.Join(userDir, "Maildir", "new"),
		filepath.Join(userDir, "Maildir", "tmp"),
	}
	for _, dir := range userSubDirs {
		userMkdir := agent.File{
			Task:  "mkdir",
			Path:  dir,
			Mode:  "0700",
			User:  "vmail",
			Group: "vmail",
		}
		req.AddFile(&userMkdir)
	}

	tmpl := filepath.Join("account", "add_user.sql")
	stmts, err := templates.SQL(tmpl, cfg.Verbose, user)
	if err != nil {
		return err
	}
	entry := agent.MySQL{
		Stmts:   stmts,
		Comment: fmt.Sprintf("add vmail account for %s", user.Email()),
	}
	req.MySQLs = append(req.MySQLs, &entry)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) ProxyDirs(domain *email.Domain) error {
	req := cfg.NewRequest()

	proxyDir := agent.GetToolsDir("data", "mailproxy", domain.Name)
	data := struct {
		Domain string
		FQDN   string
		Root   string
		MXs    []string
	}{
		Domain: domain.Name,
		FQDN:   cfg.FQDN(),
		Root:   proxyDir,
		MXs:    []string{cfg.FQDN()},
	}
	if cfg.Spambarrier != "" {
		data.MXs = append(data.MXs, "*.spambarrier.de")
	}

	// Create dir for autoconfig.xml
	autoConfigDir := filepath.Join(proxyDir, "mail")
	autoConfigMkdir := agent.File{
		Task:  "mkdir",
		Path:  autoConfigDir,
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&autoConfigMkdir)

	// create autoconfig.xml
	autoConfigContent, err := templates.Parse("mailproxy/autoconfig.xml", cfg.Verbose, data)
	if err != nil {
		return err
	}
	autoConfigFile := agent.File{
		Task:    "write",
		Path:    filepath.Join(autoConfigDir, "config-v1.1.xml"),
		Content: autoConfigContent,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
	}
	req.AddFile(&autoConfigFile)

	// Create dir for autodiscover.xml
	autoDiscoverDir := filepath.Join(proxyDir, "autodiscover")
	autoDiscoverMkdir := agent.File{
		Task:  "mkdir",
		Path:  autoDiscoverDir,
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&autoDiscoverMkdir)

	// create autodiscover.xml
	autoDiscoverContent, err := templates.Parse("mailproxy/autodiscover.xml", cfg.Verbose, data)
	if err != nil {
		return err
	}
	autoDiscoverFile := agent.File{
		Task:    "write",
		Path:    filepath.Join(autoDiscoverDir, "autodiscover.xml"),
		Content: autoDiscoverContent,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
	}
	req.AddFile(&autoDiscoverFile)

	// Create dir for and mta-sts.txt
	mtaStsDir := filepath.Join(proxyDir, ".well-known")
	mtaStsMkdir := agent.File{
		Task:  "mkdir",
		Path:  mtaStsDir,
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&mtaStsMkdir)

	// create mta-sts.txt
	mtaStsContent, err := templates.Parse("mailproxy/mta-sts.txt", cfg.Verbose, data)
	if err != nil {
		return err
	}
	mtaStsFile := agent.File{
		Task:    "write",
		Path:    filepath.Join(mtaStsDir, "mta-sts.txt"),
		Content: mtaStsContent,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
	}
	req.AddFile(&mtaStsFile)

	// Create dir for all log files
	allLogsMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetToolsDir("logs", "mailproxy"),
		Mode:  "0755",
		User:  "root",
		Group: "root",
	}
	req.AddFile(&allLogsMkdir)

	// Create dir for domain log files
	logsMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetToolsDir("logs", "mailproxy", domain.Name),
		Mode:  "0755",
		User:  "www-data",
		Group: "www-data",
	}
	req.AddFile(&logsMkdir)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) ProxyCert(domain *email.Domain) error {
	certName := "autoconfig." + domain.Name
	sanList := []string{
		"autodiscover." + domain.Name,
		"mta-sts." + domain.Name,
	}

	if err := cfg.EnsureCertificate(certName, sanList...); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) ProxyVhost(domain *email.Domain) error {
	req := cfg.NewRequest()

	certName := "autoconfig." + domain.Name
	config := struct {
		Domain   string
		SysAdmin string
		RootDir  string
		CertDir  string
		LogsDir  string
	}{
		Domain:   domain.Name,
		SysAdmin: cfg.SysAdmin,
		RootDir:  agent.GetToolsDir("data", "mailproxy", domain.Name),
		CertDir:  agent.GetToolsDir("data", "certs", certName),
		LogsDir:  agent.GetToolsDir("logs", "mailproxy", domain.Name),
	}

	tmpl, err := templates.Parse("mailproxy/vhost.conf", cfg.Verbose, config)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("14-autoconfig.%s.conf", domain.Name)
	path := agent.GetApacheEtcDir("sites-available", name)
	file := agent.File{
		Task:    "write",
		Path:    path,
		Content: tmpl,
		Service: "apache2",
	}
	req.AddFile(&file)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) DeployAccountProxy(domain *email.Domain) error {
	cnames := []string{
		"autoconfig",
		"autodiscover",
		"mta-sts",
		"imap",
		"smtp",
	}
	for _, alias := range domain.Aliases {
		cnames = append(cnames, alias)
	}

	for _, cname := range cnames {
		if status, err := cfg.SetCNAME(domain.Name, cname); err != nil {
			return err
		} else if status != "" {
			cfg.Say(status)
		}
	}

	return nil
}
