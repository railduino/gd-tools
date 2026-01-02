package config

import (
	"fmt"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/templates"
	"github.com/railduino/gd-tools/utils"
)

func (cfg *Config) DeployApache() error {
	cfg.Debug("Enter pkg/config/apache.go")

	cfg.RootDir = agent.GetApacheToolsDir()
	cfg.LogsDir = agent.GetToolsDir("logs", "apache")
	cfg.CertDir = agent.GetToolsDir("data", "certs", cfg.FQDN())

	if err := cfg.ApacheMods(); err != nil {
		return err
	}

	if err := cfg.ApacheDirs(); err != nil {
		return err
	}

	if err := cfg.ApacheIndex(); err != nil {
		return err
	}

	if err := cfg.ApacheOptions(); err != nil {
		return err
	}

	if err := cfg.ApacheDHparam("apache2"); err != nil {
		return err
	}

	emailSANs := email.GetDomainSANs()
	if err := cfg.EnsureCertificate(cfg.FQDN(), emailSANs...); err != nil {
		return err
	}

	if err := cfg.ApacheVhost(); err != nil {
		return err
	}

	cfg.AddFirewall("80/tcp")
	cfg.AddFirewall("443/tcp")
	if err := cfg.Save(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/apache.go")
	return nil
}

func (cfg *Config) ApacheMods() error {
	req := cfg.NewRequest()

	mods := []string{
		"alias.load", "alias.conf",
		"dir.load", "dir.conf",
		"env.load",
		"headers.load",
		"mime.load", "mime.conf",
		"proxy.load",
		"proxy_fcgi.load",
		"proxy_http.load",
		"rewrite.load",
		"setenvif.load", "setenvif.conf",
		"socache_shmcb.load",
		"ssl.load", "ssl.conf",
	}
	for _, mod := range mods {
		modPath := agent.GetApacheEtcDir("mods-available", mod)
		modFile := agent.File{
			Task:    "process",
			Path:    modPath,
			Service: "apache2",
		}
		req.AddFile(&modFile)
	}

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) ApacheDirs() error {
	req := cfg.NewRequest()

	rootMkdir := agent.File{
		Task:  "mkdir",
		Path:  cfg.RootDir,
		Mode:  "0750",
		User:  "www-data",
		Group: "www-data",
	}
	req.AddFile(&rootMkdir)

	logsMkdir := agent.File{
		Task:  "mkdir",
		Path:  cfg.LogsDir,
		Mode:  "0750",
		User:  "www-data",
		Group: "www-data",
	}
	req.AddFile(&logsMkdir)

	cacheMkdir := agent.File{
		Task:  "mkdir",
		Path:  agent.GetVarDir("cache", "fontconfig"),
		Mode:  "0755",
		User:  "www-data",
		Group: "www-data",
	}
	req.AddFile(&cacheMkdir)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) ApacheIndex() error {
	req := cfg.NewRequest()

	indexTmpl, err := templates.Parse("apache/index.html", cfg.Verbose, cfg)
	if err != nil {
		return err
	}
	indexPath := agent.GetApacheToolsDir("index.html")
	indexFile := agent.File{
		Task:    "write",
		Path:    indexPath,
		Content: indexTmpl,
		Mode:    "0644",
		Service: "apache2",
	}
	req.AddFile(&indexFile)

	testTmpl, err := templates.Parse("apache/test.php", cfg.Verbose, cfg)
	if err != nil {
		return err
	}
	testPath := agent.GetApacheToolsDir("test.php")
	testFile := agent.File{
		Task:    "write",
		Path:    testPath,
		Content: testTmpl,
		Mode:    "0644",
		Service: "apache2",
	}
	req.AddFile(&testFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) ApacheOptions() error {
	req := cfg.NewRequest()

	optionsTmpl, err := templates.Parse("apache/options.conf", cfg.Verbose, cfg)
	if err != nil {
		return err
	}

	optionsPath := agent.GetApacheEtcDir("conf-available", "options-ssl-apache.conf")
	optionsFile := agent.File{
		Task:    "write",
		Path:    optionsPath,
		Content: optionsTmpl,
		Service: "apache2",
	}
	req.AddFile(&optionsFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) ApacheDHparam(service string) error {
	req := cfg.NewRequest()

	dhBytes, err := utils.GenerateDHParams(2048)
	if err != nil {
		return err
	}

	dhFile := agent.File{
		Task:    "write",
		Path:    cfg.DHParamsPath(),
		Content: dhBytes,
		Mode:    "0644",
		User:    "root",
		Group:   "root",
		Service: service,
	}
	req.AddFile(&dhFile)

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) ApacheVhost() error {
	req := cfg.NewRequest()

	vhostTmpl, err := templates.Parse("apache/vhost.conf", cfg.Verbose, cfg)
	if err != nil {
		return err
	}

	vhostName := fmt.Sprintf("06-%s.conf", cfg.FQDN())
	vhostPath := agent.GetApacheEtcDir("sites-available", vhostName)
	vhostFile := agent.File{
		Task:    "write",
		Path:    vhostPath,
		Content: vhostTmpl,
		Service: "apache2",
	}
	req.AddFile(&vhostFile)

	req.AddFirewall("80/tcp")
	req.AddFirewall("443/tcp")

	req.Services = append(req.Services, "apache2")

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
