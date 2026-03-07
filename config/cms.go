package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/php"
	"github.com/railduino/gd-tools/releases"
	"github.com/railduino/gd-tools/templates"
)

const (
	CMSName  = "cms.json"
	FirstCMS = 31
	LastCMS  = 39
)

type CMS struct {
	Number     int      `json:"number"`
	Product    string   `json:"product"`
	HostName   string   `json:"host_name"`
	DomainName string   `json:"domain_name"`
	Version    string   `json:"version"`
	Aliases    []string `json:"aliases"`
	Language   string   `json:"language"`
	Region     string   `json:"region"`
	Password   string   `json:"password"`
	DirName    string   `json:"dir_name"`
	AdminName  string   `json:"admin_name"`
	AdminEmail string   `json:"admin_email"`
	AdminPswd  string   `json:"admin_pswd"`

	// for WordPress
	Salt         string `json:"salt,omitempty"`
	WpCliVersion string `json:"wp_cli_version"`

	// for MediaWiki
	Title string `json:"title,omitempty"`
}

type CMSList struct {
	Entries []*CMS `json:"entries"`
}

func (cms *CMS) Locale() string {
	return cms.Language + "_" + cms.Region
}

func (cms *CMS) FQDN() string {
	return cms.HostName + "." + cms.DomainName
}

func (cms *CMS) IsWWW() bool {
	return cms.HostName == "www"
}

func (cms *CMS) Name() string {
	if cms.IsWWW() {
		return agent.MakeDBName(cms.DomainName)
	}
	return agent.MakeDBName(cms.FQDN())
}

func (cms *CMS) ServerAlias() string {
	aliases := []string{
		cms.FQDN(),
	}
	for _, domain := range cms.Aliases {
		aliases = append(aliases, "www."+domain, domain)
	}
	sort.Strings(aliases)
	return strings.Join(aliases, " ")
}

func (cms *CMS) RootDir() string {
	return agent.GetToolsDir("data", cms.Product, cms.Name())
}

func (cms *CMS) SocketPath() string {
	name := fmt.Sprintf("php%s-%s-%s.sock", php.GetPhpVersion(), cms.Product, cms.Name())
	return filepath.Join("/run/php", name)
}

func (cms *CMS) ConfigPath() string {
	return filepath.Join(cms.RootDir(), "config.json")
}

func (cms *CMS) BaseDir(paths ...string) string {
	baseDir := filepath.Join(cms.RootDir(), cms.DirName)
	if len(paths) == 0 {
		return baseDir
	}
	return filepath.Join(append([]string{baseDir}, paths...)...)
}

func (cms *CMS) LogsDir(paths ...string) string {
	logsDir := agent.GetToolsDir("logs", cms.Product, cms.Name())
	if len(paths) == 0 {
		return logsDir
	}
	return filepath.Join(append([]string{logsDir}, paths...)...)
}

func (cms *CMS) VhostPath() string {
	name := fmt.Sprintf("%02d-%s.conf", cms.Number, cms.FQDN())
	return agent.GetApacheEtcDir("sites-available", name)
}

func (cms *CMS) HookPath() string {
	name := fmt.Sprintf("backup-pre-%s-%s", cms.Product, cms.Name())
	return agent.GetToolsDir("data", "hooks", name)
}

func (cms *CMS) CertDir() string {
	return agent.GetToolsDir("data", "certs", cms.FQDN())
}

func (cms *CMS) CertificateList() (string, []string) {
	if !cms.IsWWW() {
		return cms.FQDN(), nil
	}

	list := []string{
		cms.DomainName,
	}

	for _, domain := range cms.Aliases {
		list = append(list, "www."+domain, domain)
	}

	return cms.FQDN(), list
}

func (cms *CMS) NameList() []string {
	list := []string{
		cms.DomainName,
	}

	if cms.IsWWW() && len(cms.Aliases) > 0 {
		for _, alias := range cms.Aliases {
			list = append(list, alias)
		}
	}

	return list
}

func (cms *CMS) CronPath() string {
	name := cms.Product + "_" + cms.Name()
	return agent.GetEtcDir("cron.d", name)
}

func LoadCMSList(update *CMS) (*CMSList, error) {
	var list CMSList

	content, err := os.ReadFile(CMSName)
	if err != nil {
		if os.IsNotExist(err) {
			return &list, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", CMSName, err)
	}

	if err := json.Unmarshal(content, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", CMSName, err)
	}

	for index, _ := range list.Entries {
		entry := list.Entries[index]
		if entry.Number < FirstCMS || entry.Number > LastCMS {
			return nil, fmt.Errorf("invalid number %d for CMS %s", entry.Number, entry.FQDN())
		}
		if entry.HostName == "" {
			return nil, fmt.Errorf("found CMS without HostName")
		}
		if entry.DomainName == "" {
			return nil, fmt.Errorf("found CMS without DomainName")
		}
		if entry.Product == "" {
			return nil, fmt.Errorf("found CMS without Product")
		}

		if entry.DirName == "" {
			return nil, fmt.Errorf("missing DirName for CMS %s", entry.FQDN())
		}

		if update != nil && update.Password != "" {
			entry.Password = update.Password
		}
		if entry.Password == "" {
			return nil, fmt.Errorf("missing Password for CMS %s", entry.FQDN())
		}

		if update != nil && update.AdminName != "" {
			entry.AdminName = update.AdminName
		}
		if entry.AdminName == "" {
			return nil, fmt.Errorf("missing AdminName for CMS %s", entry.FQDN())
		}

		if update != nil && update.AdminEmail != "" {
			entry.AdminEmail = update.AdminEmail
		}
		if entry.AdminEmail == "" {
			return nil, fmt.Errorf("missing AdminEmail for CMS %s", entry.FQDN())
		}

		if update != nil && update.AdminPswd != "" {
			entry.AdminPswd = update.AdminPswd
		}
		if entry.AdminPswd == "" {
			return nil, fmt.Errorf("missing AdminPswd for CMS %s", entry.FQDN())
		}

		// place CMS specific code here
		switch entry.Product {
		case "wordpress":
			if update != nil && update.Salt != "" {
				entry.Salt = update.Salt
			}
			if entry.Salt == "" {
				return nil, fmt.Errorf("missing Salt for WordPress %s", entry.FQDN())
			}
		case "mediawiki":
			if update != nil && update.Title != "" {
				entry.Title = update.Title
			}

		}
	}

	if err := list.Save(); err != nil {
		return nil, err
	}

	if update != nil {
		return nil, nil
	}

	return &list, nil
}

func (list *CMSList) Save() error {
	sort.Slice(list.Entries, func(i, j int) bool {
		return list.Entries[i].Number < list.Entries[j].Number
	})

	content, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", CMSName, err)
	}

	existing, err := os.ReadFile(CMSName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(CMSName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", CMSName, err)
	}

	return nil
}

func (cfg *Config) DeployCMS() error {
	cms := cfg.CMS
	if cms == nil {
		return fmt.Errorf("missing CMS pointer")
	}
	cfg.Debugf("Enter onfig/cms.go (%s: %s)", cms.FQDN(), cms.Product)

	if cms.FQDN() == cfg.FQDN() {
		return fmt.Errorf("cannot use the server name for CMS")
	}

	if err := cfg.CMSDownload(cms); err != nil {
		return err
	}

	if err := cfg.CMSExtract(cms); err != nil {
		return err
	}

	if err := cfg.CMSLogsDir(cms); err != nil {
		return err
	}

	if err := cfg.CMS_SQL(cms); err != nil {
		return err
	}

	if err := cfg.CMSConfig(cms); err != nil {
		return err
	}

	if err := cfg.CMSBackupHook(cms); err != nil {
		return err
	}

	if err := cfg.CMS_DNS(cms); err != nil {
		return err
	}

	// place CMS specific code here
	switch cms.Product {
	case "wordpress":
		if err := cfg.WordPressExtras(cms); err != nil {
			return err
		}
	case "mediawiki":
		if err := cfg.MediaWikiInstall(cms); err != nil {
			return err
		}
	}

	cfg.Debugf("Leave onfig/cms.go (%s: %s)", cms.FQDN(), cms.Product)
	return nil
}

func (cfg *Config) CMSDownload(cms *CMS) error {
	req := cfg.NewRequest()

	cat, err := releases.Load(cfg.Verbose)
	if err != nil {
		return err
	}
	_, rel, err := cat.Get(cms.Product, cms.Version)
	if err != nil {
		return err
	}
	if rel.Download.Directory == "" {
		return fmt.Errorf("missing Directory in %s download", cms.Product)
	}

	req.Downloads = append(req.Downloads, &rel.Download)

	if cms.Product == "wordpress" {
		_, rel, err := cat.Get("wp-cli", cms.WpCliVersion)
		if err != nil {
			return err
		}
		if rel.Download.Binary == "" {
			return fmt.Errorf("missing Binary in wp-cli download")
		}
		req.Downloads = append(req.Downloads, &rel.Download)
	}

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) CMSExtract(cms *CMS) error {
	req := cfg.NewRequest()

	cat, err := releases.Load(cfg.Verbose)
	if err != nil {
		return err
	}
	_, rel, err := cat.Get(cms.Product, cms.Version)
	if err != nil {
		return err
	}

	downloadDir := agent.GetDownloadsDir(rel.Download.Filename)
	extract := agent.File{
		Task:   "extract",
		Path:   downloadDir,
		Target: cms.RootDir(),
		Mode:   "0755",
		User:   "root",
		Group:  "root",
	}
	req.AddFile(&extract)

	switch cms.Product {
	case "wbce":
		writableDirs := []string{
			"pages",
			"media",
			"var",
			"languages",
			"templates",
			"modules",
			"temp",
		}
		for _, subDir := range writableDirs {
			writableTask := agent.File{
				Task:  "process",
				Path:  filepath.Join(cms.BaseDir(), subDir),
				User:  "www-data",
				Group: "www-data",
			}
			req.AddFile(&writableTask)
		}

	case "wordpress":
		writableTask := agent.File{
			Task:  "process",
			Path:  cms.BaseDir(),
			User:  "www-data",
			Group: "www-data",
		}
		req.AddFile(&writableTask)
	}

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) CMSLogsDir(cms *CMS) error {
	req := cfg.NewRequest()

	logsMkdir := agent.File{
		Task:  "mkdir",
		Path:  cms.LogsDir(),
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

func (cfg *Config) CMS_SQL(cms *CMS) error {
	req := cfg.NewRequest()

	sqlTmpl := filepath.Join(cms.Product, "create.sql")
	sqlStmts, err := templates.SQL(sqlTmpl, cfg.Verbose, cms)
	if err != nil {
		return err
	}

	sql := agent.MySQL{
		Stmts:   sqlStmts,
		Comment: fmt.Sprintf("create %s (%s) tables", cms.Product, cms.FQDN()),
	}
	req.MySQLs = append(req.MySQLs, &sql)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) CMSBackupHook(cms *CMS) error {
	req := cfg.NewRequest()

	hookPath := filepath.Join(cms.Product, "backup")
	hookContent, err := templates.Parse(hookPath, cfg.Verbose, cms)
	if err != nil {
		return err
	}

	hookFile := agent.File{
		Task:    "write",
		Path:    cms.HookPath(),
		Content: hookContent,
		Mode:    "0500",
		User:    "root",
		Group:   "root",
	}
	req.AddFile(&hookFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) CMSConfig(cms *CMS) error {
	req := cfg.NewRequest()

	var confTmpl []byte
	var confPath, confUser string
	var err error
	switch cms.Product {
	case "wbce":
		confTmpl, err = templates.Parse("wbce/config.php", cfg.Verbose, cms)
		confPath = cms.BaseDir("config.php")
		confUser = "www-data"
	case "wordpress":
		confTmpl, err = templates.Parse("wordpress/wp-config.php", cfg.Verbose, cms)
		confPath = cms.BaseDir("wp-config.php")
		confUser = "www-data"
	case "mediawiki":
		// no template here; LocalSettings.php is created by the installer on Prod
		confTmpl = nil
	default:
		return fmt.Errorf("unknown CMS product '%s'", cms.Product)
	}
	if err != nil {
		return err
	}

	if len(confTmpl) > 0 {
		confFile := agent.File{
			Task:    "write",
			Path:    confPath,
			Content: confTmpl,
			Mode:    "0644",
			User:    confUser,
			Group:   confUser,
			Service: "apache2",
		}
		req.AddFile(&confFile)
	}

	poolName := filepath.Join(cms.Product, "php-fpm-pool.conf")
	poolTmpl, err := templates.Parse(poolName, cfg.Verbose, cms)
	if err != nil {
		return err
	}

	poolPath := php.GetPhpFpmPoolPath(cms.Number, cms.Product+"-"+cms.Name())
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

func (cfg *Config) CMS_DNS(cms *CMS) error {
	// get certificates for all possible names via DNS-01
	fqdnCert, sanCerts := cms.CertificateList()
	if err := cfg.EnsureCertificate(fqdnCert, sanCerts...); err != nil {
		return err
	}

	// setup vhost with all aliases, www and domain
	if err := cfg.CMSSetupVhost(cms); err != nil {
		return err
	}

	if cfg.SkipDNS {
		cfg.Sayf("skipping dns-update for %s", cms.FQDN())
		return nil
	}

	for _, name := range cms.NameList() {
		// install all www entries as CNAME records
		// The NameList contains at least the FQDN of the CMS
		if status, err := cfg.SetCNAME(name, cms.HostName); err != nil {
			return err
		} else if status != "" {
			cfg.Say(status)
		}

		// for non-www servers (e.g. demo.example.com) there is nothing more to do
		if !cms.IsWWW() {
			continue
		}

		// install the domain A and AAAA records (those cannot be CNAME)
		if cfg.IPv4Addr != "" {
			if status, err := cfg.SetA(name, "@", cfg.IPv4Addr); err != nil {
				return err
			} else if status != "" {
				cfg.Say(status)
			}
		}
		if cfg.IPv6Addr != "" {
			if status, err := cfg.SetAAAA(name, "@", cfg.IPv6Addr); err != nil {
				return err
			} else if status != "" {
				cfg.Say(status)
			}
		}
	}

	return nil
}

func (cfg *Config) CMSSetupVhost(cms *CMS) error {
	req := cfg.NewRequest()

	vhostTmpl := filepath.Join(cms.Product, "vhost.conf")
	vhostContent, err := templates.Parse(vhostTmpl, cfg.Verbose, cms)
	if err != nil {
		return err
	}

	vhostFile := agent.File{
		Task:    "write",
		Path:    cms.VhostPath(),
		Content: vhostContent,
	}
	req.AddFile(&vhostFile)
	req.AddService("apache2")

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
