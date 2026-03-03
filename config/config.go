package config

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/php"
	"github.com/railduino/gd-tools/templates"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

const (
	ConfigName    = "config.json"
	ConfigTimeout = 10
)

type Config struct {
	Version string `json:"version"` // e.g. v1.0

	Verbose   bool             `json:"-"`
	Quiet     bool             `json:"-"`
	Dry       bool             `json:"-"`
	Force     bool             `json:"-"`
	Delete    bool             `json:"-"`
	SkipDNS   bool             `json:"-"`
	SkipMX    bool             `json:"-"`
	Port      string           `json:"-"`
	Conn      *tls.Conn        `json:"-"`
	Req       *agent.Request   `json:"-"`
	CmdEnv    []string         `json:"-"`
	Mailer    *Mailer          `json:"-"`
	DKIMs     []string         `json:"-"`
	Postfix   *Postfix         `json:"-"`
	Nextcloud *agent.Nextcloud `json:"-"`
	OCIS      *OCIS            `json:"-"`
	CMS       *CMS             `json:"-"`
	RustDesk  *agent.RustDesk  `json:"-"`
	Password  string           `json:"-"`
	RootDir   string           `json:"-"`
	BaseDir   string           `json:"-"`
	LogsDir   string           `json:"-"`
	CertDir   string           `json:"-"`

	Company  string `json:"company"`   // company name, used e.g. for Webmail
	SysAdmin string `json:"sys_admin"` // try to read from ~/.gitconfig
	TimeZone string `json:"time_zone"` // e.g. Europe/Berlin
	Language string `json:"language"`  // e.g. de
	Region   string `json:"region"`    // e.g. DE
	RegTTL   int    `json:"reg_ttl"`   // DNS TTL for static names, default is 3600
	HelpURL  string `json:"help_url"`  // Support URL

	HostName   string         `json:"host_name"`   // host part of FQDN
	DomainName string         `json:"domain_name"` // domain part of FQDN
	IPv4Addr   string         `json:"ipv4_addr"`   // must be set manually
	IPv6Addr   string         `json:"ipv6_addr"`   // must be set manually
	DMARC      string         `json:"dmarc"`       // DMARC value (p=none, rua=mailto:<admin>)
	SwapSize   string         `json:"swap_size"`   // e.g. 500M or 2G, or 0
	Timeout    int            `json:"timeout"`     // connection timeout, default 10 seconds
	Packages   []string       `json:"packages"`    // required DEB packages
	Mounts     []*agent.Mount `json:"mounts"`      // mounted filesystem (which can grow)
	Firewall   []string       `json:"firewall"`    // ufw ports to open, e.g. 22/tcp

	Spambarrier string `json:"spambarrier"` // spambarrier.de API key (see https://www.spambarrier.de/)
	UbuntuPro   string `json:"ubuntu_pro"`  // Ubuntu Pro Key (see https://ubuntu.com/pro)

	HetznerToken string `json:"hetzner_token"` // Token for Hetzner Cloud API
	IonosToken   string `json:"ionos_token"`   // Token for IONOS Cloud API
	// add more provider token here
}

var (
	FlagDry = &cli.BoolFlag{
		Name:  "dry",
		Usage: "display commands instead of executing them",
	}
	FlagVerbose = &cli.BoolFlag{
		Name:  "verbose",
		Usage: "print extended debugging information",
	}
	FlagForce = &cli.BoolFlag{
		Name:  "force",
		Usage: "enforce action even if not necessary",
	}
	FlagSkipDNS = &cli.BoolFlag{
		Name:  "skip-dns",
		Usage: "do not execute DNS installs or updates",
	}
	FlagSkipMX = &cli.BoolFlag{
		Name:  "skip-mx",
		Usage: "do not touch the domain MX record",
	}
	FlagDelete = &cli.BoolFlag{
		Name:  "delete",
		Usage: "delete the given resource - may the --force be with you",
	}
	FlagPort = &cli.StringFlag{
		Name:  "port",
		Usage: "TCP port for dev/prod communication",
		Value: agent.DefaultPort,
	}
)

func (cfg *Config) Locale() string {
	return cfg.Language + "_" + cfg.Region
}

func (cfg *Config) FQDN() string {
	return cfg.HostName + "." + cfg.DomainName
}

func (cfg *Config) DHParamsPath() string {
	return agent.GetEtcDir("apache2", utils.DHParamsName)
}

func (cfg *Config) RootUser() string {
	return "root@" + cfg.FQDN()
}

func (cfg *Config) FQDNdot() string {
	return cfg.HostName + "." + cfg.DomainName + "."
}

func (cfg *Config) DotFQDN() string {
	return "." + cfg.HostName + "." + cfg.DomainName
}

func (cfg *Config) RsyncFlags() string {
	if cfg.Verbose {
		return "-avz"
	}
	return "-avzq"
}

func ReadConfigPlus(c *cli.Context) (*Config, *agent.Request, error) {
	cfg, err := ReadConfig(c)
	if err != nil {
		return nil, nil, err
	}

	timeout := cfg.Timeout
	tmpTime := c.Int("timeout")
	if tmpTime != 0 && tmpTime != timeout {
		timeout = tmpTime
	}

	cfg.Conn, err = agent.ConnectToAgent(cfg.FQDN(), timeout, cfg.Verbose)
	if err != nil {
		return cfg, nil, err
	}

	req := cfg.NewRequest()

	return cfg, req, nil
}

func (cfg *Config) Close() {
	if cfg != nil && cfg.Conn != nil {
		cfg.Conn.Close()
		cfg.Conn = nil
	}
}

func ReadConfig(c *cli.Context) (*Config, error) {
	content, err := os.ReadFile(ConfigName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("missing %s - are we in the correct dir?", ConfigName)
		}
		return nil, fmt.Errorf("failed to read %s: %w", ConfigName, err)
	}

	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", ConfigName, err)
	}

	if c != nil {
		cfg.Verbose = c.Bool("verbose")
		cfg.Dry = c.Bool("dry")
		cfg.Force = c.Bool("force")
		cfg.Delete = c.Bool("delete")
		cfg.SkipDNS = c.Bool("skip-dns")
		cfg.SkipMX = c.Bool("skip-mx")
		cfg.Port = c.String("port")
		agent.SetAgentPort(cfg.Port)
	}

	if cfg.HostName == "" {
		return nil, fmt.Errorf("missing HostName in config.json")
	}
	if cfg.DomainName == "" {
		return nil, fmt.Errorf("missing DomainName in config.json")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = ConfigTimeout
	}

	basics, err := utils.GetBasics()
	if err != nil {
		return nil, err
	}
	if cfg.Company == "" {
		cfg.Company = basics.Company
	}
	if cfg.SysAdmin == "" {
		cfg.SysAdmin = basics.SysAdmin
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = basics.TimeZone
	}
	if cfg.Language == "" {
		cfg.Language = basics.Language
	}
	if cfg.Region == "" {
		cfg.Region = basics.Region
	}
	if cfg.RegTTL == 0 {
		cfg.RegTTL = basics.RegTTL
	}
	if cfg.DMARC == "" {
		cfg.DMARC = fmt.Sprintf("v=DMARC1; p=none; rua=mailto:%s", cfg.SysAdmin)
	}

	packagesPath := filepath.Join("assets", "packages.txt")
	packages, err := templates.Lines(packagesPath, "#", cfg.Verbose, php.GetTemplateData())
	if err != nil {
		return nil, err
	}
	if equal := StringSlicesEqual(packages, cfg.Packages); !equal {
		cfg.Sayf("packages have changed (%d -> %d)", len(cfg.Packages), len(packages))
		cfg.Packages = packages
	}

	if err := os.MkdirAll(ACME_Cert_Dir, 0755); err != nil {
		return nil, err
	}

	if _, _, err := utils.GetRSAKeyPair(cfg.FQDN()); err != nil {
		return nil, err
	}

	cfg.AddFirewall("22/tcp")
	cfg.AddFirewall(agent.GetAgentPort() + "/tcp")

	ips, err := net.LookupIP(cfg.FQDN())
	if err != nil {
		return nil, err
	}

	for _, ip := range ips {
		cfg.Debugf("IP Address: %s", ip.String())
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			if cfg.IPv4Addr != ip.String() {
				if cfg.IPv4Addr == "" {
					cfg.IPv4Addr = ip.String()
					cfg.Sayf("IPv4Addr defaults to %s", cfg.IPv4Addr)
				} else {
					cfg.Debugf("IPv4Addr was changed to %s", cfg.IPv4Addr)
				}
			}
			break
		}
	}

	for _, ip := range ips {
		if ip.To16() != nil && ip.To4() == nil {
			if cfg.IPv6Addr != ip.String() {
				if cfg.IPv6Addr == "" {
					cfg.IPv6Addr = ip.String()
					cfg.Sayf("IPv6Addr defaults to %s", cfg.IPv6Addr)
				} else {
					cfg.Debugf("IPv6Addr was changed to %s", cfg.IPv6Addr)
				}
			}
			break
		}
	}

	if err := cfg.Save(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) NewRequest() *agent.Request {
	return &agent.Request{
		Version:  agent.ProtocolVersion,
		Conn:     cfg.Conn,
		Verbose:  cfg.Verbose,
		Quiet:    cfg.Quiet,
		Dry:      cfg.Dry,
		Language: cfg.Language,
		Region:   cfg.Region,
	}
}

func (cfg *Config) Save() error {
	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", ConfigName, err)
	}

	existing, err := os.ReadFile(ConfigName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(ConfigName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", ConfigName, err)
	}
	cfg.Say("config.json has been updated - please check")

	return nil
}

func (cfg *Config) Say(args ...any) bool {
	if cfg.Quiet {
		return cfg.Dry
	}
	prefix := "[run]"
	if cfg.Dry {
		prefix = "[dry]"
	}

	for _, arg := range args {
		switch v := arg.(type) {
		case []string:
			for _, line := range v {
				fmt.Println(prefix, line)
			}
		default:
			fmt.Println(prefix, fmt.Sprint(arg))
		}
	}

	return cfg.Dry
}

func (cfg *Config) Sayf(format string, args ...any) bool {
	if cfg.Quiet {
		return cfg.Dry
	}
	prefix := "[run]"
	if cfg.Dry {
		prefix = "[dry]"
	}
	fmt.Println(prefix, fmt.Sprintf(format, args...))
	return cfg.Dry
}

func (cfg *Config) Debug(args ...any) bool {
	if !cfg.Verbose {
		return cfg.Dry
	}

	prefix := "[run] ##########"
	if cfg.Dry {
		prefix = "[dry] ##########"
	}

	for _, arg := range args {
		switch v := arg.(type) {
		case []string:
			for _, line := range v {
				fmt.Println(prefix, line)
			}
		default:
			fmt.Println(prefix, fmt.Sprint(arg))
		}
	}

	return cfg.Dry
}

func (cfg *Config) Debugf(format string, args ...any) bool {
	if !cfg.Verbose {
		return cfg.Dry
	}

	prefix := "[run] ##########"
	if cfg.Dry {
		prefix = "[dry] ##########"
	}
	fmt.Println(prefix, fmt.Sprintf(format, args...))
	return cfg.Dry
}

func (cfg *Config) AddFirewall(port string) {
	for _, open := range cfg.Firewall {
		if port == open {
			return
		}
	}

	cfg.Firewall = append(cfg.Firewall, port)
	sort.Strings(cfg.Firewall)
}

func StringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]int)
	for _, s := range a {
		m[s]++
	}
	for _, s := range b {
		if m[s] == 0 {
			return false
		}
		m[s]--
	}
	return true
}
