package config

import (
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
	"github.com/railduino/gd-tools/utils"
)

type Postfix struct {
	HostName  string
	Password  string
	CertDir   string
	MailPath  string
	VmailUID  string
	VmailGID  string
	MilterIn  string
	MilterOut string
}

func (cfg *Config) DeployPostfix() error {
	cfg.Debug("Enter pkg/config/postfix.go")

	// the Mailer is defined in pkg/config/accounts.go
	mailer, err := LoadMailer()
	if err != nil {
		return err
	}
	cfg.Mailer = mailer

	password, err := utils.FetchPassword(20, "vmail", "db_password")
	if err != nil {
		return err
	}

	cfg.Postfix = &Postfix{
		HostName: cfg.FQDN(),
		Password: password,
		CertDir:  agent.GetToolsDir("data", "certs", cfg.FQDN()),
		VmailUID: mailer.VmailUID,
		VmailGID: mailer.VmailGID,
		MailPath: mailer.MailPath,
	}

	if cfg.Spambarrier != "" {
		cfg.Postfix.MilterIn = ""
	} else {
		cfg.Postfix.MilterIn = "inet:localhost:11332"
	}

	if err := cfg.PostfixTables(); err != nil {
		return err
	}

	if err := cfg.PostfixMain(); err != nil {
		return err
	}

	if err := cfg.PostfixMaster(); err != nil {
		return err
	}

	cfg.AddFirewall("25/tcp")
	cfg.AddFirewall("465/tcp")
	cfg.AddFirewall("587/tcp")
	if err := cfg.Save(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/postfix.go")
	return nil
}

func (cfg *Config) PostfixTables() error {
	req := cfg.NewRequest()

	files := []string{
		"mailbox-domains.cf",
		"mailbox-maps.cf",
		"alias-maps.cf",
	}

	for _, name := range files {
		tmpl := filepath.Join("postfix", name)
		content, err := templates.Parse(tmpl, cfg.Verbose, cfg.Postfix)
		if err != nil {
			return err
		}

		file := agent.File{
			Task:    "write",
			Path:    agent.GetEtcDir("postfix", name),
			Content: content,
			Mode:    "0644",
			Service: "postfix",
		}
		req.AddFile(&file)
	}

	tmpl := filepath.Join("postfix", "create_tables.sql")
	stmts, err := templates.SQL(tmpl, cfg.Verbose, cfg.Postfix)
	if err != nil {
		return err
	}
	entry := agent.MySQL{
		Stmts:   stmts,
		Comment: "create postfix (vmail) tables",
	}
	req.MySQLs = append(req.MySQLs, &entry)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) PostfixMain() error {
	req := cfg.NewRequest()

	tmpl := filepath.Join("postfix", "main.cf")
	content, err := templates.Parse(tmpl, cfg.Verbose, cfg.Postfix)
	if err != nil {
		return err
	}

	file := agent.File{
		Task:    "write",
		Path:    agent.GetEtcDir("postfix", "main.cf"),
		Content: content,
		Backup:  true,
		Mode:    "0644",
		Service: "postfix",
	}
	req.AddFile(&file)

	req.AddFirewall("25/tcp")
	req.AddFirewall("465/tcp")
	req.AddFirewall("587/tcp")

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) PostfixMaster() error {
	req := cfg.NewRequest()

	tmpl := filepath.Join("postfix", "master.cf")
	content, err := templates.Parse(tmpl, cfg.Verbose, cfg.Postfix)
	if err != nil {
		return err
	}

	file := agent.File{
		Task:    "write",
		Path:    agent.GetEtcDir("postfix", "master.cf"),
		Content: content,
		Backup:  true,
		Mode:    "0644",
		Service: "postfix",
	}
	req.AddFile(&file)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
