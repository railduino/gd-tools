package config

import (
	"fmt"

	"github.com/railduino/gd-tools/agent"
)

// MediaWikiInstall triggers the MediaWiki CLI installer on Prod (via agent).
// MediaWiki is installed idempotently by checking LocalSettings.php on the target.
func (cfg *Config) MediaWikiInstall(cms *CMS) error {
	if cms == nil {
		return fmt.Errorf("missing CMS pointer in MediaWikiInstall()")
	}
	if cms.Product != "mediawiki" {
		return fmt.Errorf("invalid product for MediaWikiInstall(): %s", cms.Product)
	}

	req := cfg.NewRequest()

	title := cms.Title
	if title == "" {
		title = cms.FQDN()
	}

	// BaseDir is the extracted MediaWiki directory (contains maintenance/install.php)
	task := &agent.MediaWiki{
		Name:       cms.Name(),
		BaseDir:    cms.BaseDir(),
		FQDN:       cms.FQDN(),
		Title:      title,
		ScriptPath: "/", // wiki served at vhost root; adjust later if you want /wiki
		Language:   cms.Language,

		AdminName:  cms.AdminName,
		AdminEmail: cms.AdminEmail,
		AdminPswd:  cms.AdminPswd,

		DBName:     cms.Name(),
		DBUser:     cms.Name(),
		DBPassword: cms.Password,
	}

	req.MediaWiki = task
	req.MWConf = agent.MW_Install // optional; handler also installs if empty

	if err := req.Send(); err != nil {
		return fmt.Errorf("mediawiki install failed (%s): %w", cms.FQDN(), err)
	}

	return nil
}
