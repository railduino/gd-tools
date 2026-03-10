package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/templates"
	"github.com/railduino/gd-tools/utils"
)

const (
	RoutingName = "routing.json"
)

type Postfix struct {
	HostName   string
	DomainName string
	Password   string
	CertDir    string
	MailPath   string
	VmailUID   string
	VmailGID   string
	MilterIn   string
	MilterOut  string
}

type RelayDomain struct {
	Domain string `json:"domain"`
	Relay  string `json:"relay"`
}

type PolicyDomain struct {
	Domain string `json:"domain"`
	Policy string `json:"policy"`
}

type Routing struct {
	RelayDomains  []RelayDomain  `json:"relay_domains"`
	PolicyDomains []PolicyDomain `json:"policy_domains"`
}

func (cfg *Config) DeployPostfix() error {
	cfg.Debug("Enter pkg/config/postfix.go")

	// the Mailer is defined in config/accounts.go
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
		HostName:   cfg.FQDN(),
		DomainName: cfg.DomainName,
		Password:   password,
		CertDir:    agent.GetToolsDir("data", "certs", cfg.FQDN()),
		VmailUID:   mailer.VmailUID,
		VmailGID:   mailer.VmailGID,
		MailPath:   mailer.MailPath,
	}

	if cfg.Spambarrier != "" {
		cfg.Postfix.MilterIn = ""
	} else {
		cfg.Postfix.MilterIn = "inet:localhost:11332"
	}

	if err := cfg.PostfixTables(); err != nil {
		return err
	}

	if err := cfg.PostfixSASL(); err != nil {
		return err
	}

	if err := cfg.PostfixMaps(); err != nil {
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

func (cfg *Config) PostfixSASL() error {
	req := cfg.NewRequest()

	// collect the provider credentials
	var lines []string

	line, err := email.BrevoSASL()
	if err != nil {
		return err
	}
	if line != "" {
		lines = append(lines, line)
	}
	// add more providers here

	sort.Strings(lines)
	content := strings.Join(lines, "\n") + "\n"
	file := agent.File{
		Task:    "postmap",
		Path:    agent.GetEtcDir("postfix", "sasl_passwd"),
		Content: []byte(content),
		Mode:    "0600",
		Service: "postfix",
	}
	req.AddFile(&file)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) PostfixMaps() error {
	req := cfg.NewRequest()

	content, err := os.ReadFile(RoutingName)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", RoutingName, err)
	}

	var routing Routing
	if err := json.Unmarshal(content, &routing); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", RoutingName, err)
	}

	var transportLines []string
	for _, transport := range routing.RelayDomains {
		switch transport.Relay {
		case "brevo":
			server, port := email.BrevoTarget()
			if err != nil {
				return err
			}
			line := fmt.Sprintf("%s smtp:[%s]:%d", transport.Domain, server, port)
			transportLines = append(transportLines, line)
		}
		// add more providers here
	}
	sort.Strings(transportLines)

	transportData := strings.Join(transportLines, "\n") + "\n"
	transportFile := agent.File{
		Task:    "postmap",
		Path:    agent.GetEtcDir("postfix", "transport"),
		Content: []byte(transportData),
		Mode:    "0600",
		Service: "postfix",
	}
	req.AddFile(&transportFile)

	var policyLines []string
	for _, policy := range routing.PolicyDomains {
		line := fmt.Sprintf("%s %s", policy.Domain, policy.Policy)
		policyLines = append(policyLines, line)
	}
	sort.Strings(policyLines)

	policyData := strings.Join(policyLines, "\n") + "\n"
	policyFile := agent.File{
		Task:    "postmap",
		Path:    agent.GetEtcDir("postfix", "tls_policy"),
		Content: []byte(policyData),
		Mode:    "0600",
		Service: "postfix",
	}
	req.AddFile(&policyFile)

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
