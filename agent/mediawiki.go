package agent

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	MW_Install   = "install"
	MW_Installer = "maintenance/install.php"
)

// MediaWiki contains the parameters required to run the MediaWiki CLI installer on Prod.
// It is intentionally minimal and does not attempt to model MediaWiki internals.
type MediaWiki struct {
	// Identity
	Name    string `json:"name"`     // instance name (e.g. cms.Name())
	BaseDir string `json:"base_dir"` // directory where MediaWiki is installed (contains maintenance/)
	FQDN    string `json:"fqdn"`     // public name, used for --server

	// Site
	Title      string `json:"title,omitempty"`       // defaults to FQDN
	ScriptPath string `json:"script_path,omitempty"` // defaults to "/"
	Language   string `json:"language,omitempty"`    // e.g. "de"

	// Admin
	AdminName  string `json:"admin_name"`            // e.g. "WikiAdmin"
	AdminEmail string `json:"admin_email,omitempty"` // optional
	AdminPswd  string `json:"admin_pswd"`            // required

	// Database (MySQL/MariaDB)
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
}

// LocalSettingsPath is the expected config artefact after a successful install.
func (mw *MediaWiki) LocalSettingsPath() string {
	return filepath.Join(mw.BaseDir, "LocalSettings.php")
}

// IsInstalled checks for an existing LocalSettings.php.
// It is a pragmatic idempotency check, similar to Nextcloud's config.php check.
func (mw *MediaWiki) IsInstalled() bool {
	info, err := os.Stat(mw.LocalSettingsPath())
	return err == nil && info.Size() > 0
}

// MediaWikiTest checks if there is work to be done
func MediaWikiTest(req *Request) bool {
	return req != nil && req.MediaWiki != nil
}

func MediaWikiHandler(req *Request, resp *Response) error {
	if req == nil || req.MediaWiki == nil || resp == nil {
		return nil
	}

	// Optional command routing (future-proof, mirrors Nextcloud style)
	if req.MWConf != "" && req.MWConf != MW_Install {
		return fmt.Errorf("unknown MediaWiki command: %s", req.MWConf)
	}

	return req.MediaWiki.Install(resp)
}

// Install runs the MediaWiki CLI installer if LocalSettings.php is missing.
func (mw *MediaWiki) Install(resp *Response) error {
	if resp == nil {
		return fmt.Errorf("missing response in MediaWiki.Install()")
	}

	if mw.IsInstalled() {
		resp.Sayf("✅ MediaWiki '%s' installed", mw.Name)
		return nil
	}

	if mw.BaseDir == "" {
		return fmt.Errorf("missing BaseDir for MediaWiki '%s'", mw.Name)
	}
	if mw.FQDN == "" {
		return fmt.Errorf("missing FQDN for MediaWiki '%s'", mw.Name)
	}
	if mw.AdminName == "" {
		return fmt.Errorf("missing AdminName for MediaWiki '%s'", mw.Name)
	}
	if mw.AdminPswd == "" {
		return fmt.Errorf("missing AdminPswd for MediaWiki '%s'", mw.Name)
	}
	if mw.DBName == "" || mw.DBUser == "" || mw.DBPassword == "" {
		return fmt.Errorf("missing DB settings for MediaWiki '%s'", mw.Name)
	}

	// Defaults
	title := mw.Title
	if title == "" {
		title = mw.FQDN
	}
	scriptPath := mw.ScriptPath
	if scriptPath == "" {
		scriptPath = "/"
	}

	resp.Sayf("installing MediaWiki '%s'", mw.Name)

	// MediaWiki will create LocalSettings.php and may touch cache directories.
	// Ensure ownership is correct (same approach as Nextcloud).
	if err := ChownR(mw.BaseDir, "www-data:www-data"); err != nil {
		return fmt.Errorf("chown failed for %s: %w", mw.BaseDir, err)
	}
	resp.Sayf("MediaWiki '%s' belongs to www-data", mw.Name)

	// Build install args.
	// install.php expects: <WikiName> <AdminUser> as trailing args.
	args := []string{
		MW_Installer,
		"--dbtype", "mysql",
		"--dbserver", "localhost",
		"--dbname", mw.DBName,
		"--dbuser", mw.DBUser,
		"--dbpass", mw.DBPassword,

		"--server", "https://" + mw.FQDN,
		"--scriptpath", scriptPath,

		"--pass", mw.AdminPswd,
	}

	if mw.Language != "" {
		args = append(args, "--lang", mw.Language)
	}
	if mw.AdminEmail != "" {
		args = append(args, "--adminemail", mw.AdminEmail)
	}

	// Positional arguments
	args = append(args, title, mw.AdminName)

	resp.Sayf("running MediaWiki installer for '%s' - please be patient ...", mw.Name)
	if err := mw.RunPHPAsWWWData(resp, args...); err != nil {
		return err
	}

	// Verify artefact
	if !mw.IsInstalled() {
		return fmt.Errorf("MediaWiki '%s' install did not produce LocalSettings.php", mw.Name)
	}

	resp.Sayf("✅ MediaWiki '%s' installed", mw.Name)
	return nil
}

// RunPHPAsWWWData runs php with the given args as user www-data in mw.BaseDir.
// Output is captured and forwarded to resp similar to Nextcloud.RunOCC().
func (mw *MediaWiki) RunPHPAsWWWData(resp *Response, args ...string) error {
	if resp == nil {
		return fmt.Errorf("missing response in RunPHPAsWWWData()")
	}

	usr, err := user.Lookup("www-data")
	if err != nil {
		return fmt.Errorf("user lookup for www-data failed: %w", err)
	}
	uid, _ := strconv.Atoi(usr.Uid)
	gid, _ := strconv.Atoi(usr.Gid)

	cmd := exec.Command("/usr/bin/php", args...)
	cmd.Dir = mw.BaseDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("MediaWiki installer failed: %w\n%s", err, output.String())
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			resp.Say(line)
		}
	}

	return nil
}
