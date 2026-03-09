package install

import (
	"fmt"
	"os"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/templates"
	"github.com/urfave/cli/v2"
)

const (
	HetznerTempName = "hetzner.api"
)

var Describe = `The install command turns an innocent server into a gd-tools production server.

 - the gd-tools binary is copied to /usr/local/bin (or $GD_TOOLS_BIN_DIR)
 - a systemd service is installed to run the program as demon
 - the demon is started immediately and the party should begin

Please note that this binary depends on the CA used for deployment.
It uses mTLS and the server certificate is baked into the program
to avoid tampering with it.  If the (Dev) CA is changed, then the
program needs to be updated, too.  At least to accept the updated
client certificate.`

var Command = &cli.Command{
	Name:        "install",
	Usage:       "Create or update a gd-tools production server",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if err := cfg.SetupCA(); err != nil {
		return err
	}

	etcTools := agent.GetEtcDir("gd-tools")
	mkdirCmd := "install -o root -g root -m 0700 -d " + etcTools
	if err := cfg.RemoteCmd(mkdirCmd); err != nil {
		return fmt.Errorf("failed to install %s: %w", etcTools, err)
	}

	if _, err := cfg.LocalCommand(
		"rsync",
		cfg.RsyncFlags(),
		"--chown=root:root",
		"--chmod=0600",
		"CA/server.crt",
		"CA/server.key",
		"CA/ca.crt",
		cfg.RootUser()+":"+etcTools,
	); err != nil {
		return fmt.Errorf("failed to push mTLS certs: %w", err)
	}

	gdProgs := []string{
		"gd-tools",
	}
	for _, prog := range gdProgs {
		progPath := agent.GetBinDir(prog)
		if _, err := cfg.LocalCommand(
			"rsync",
			cfg.RsyncFlags(),
			"--chown=root:root",
			"--chmod=0500",
			progPath,
			cfg.RootUser()+":"+progPath,
		); err != nil {
			return fmt.Errorf("failed to install %s: %w", progPath, err)
		}
	}

	service := "gd-tools.service"
	data := struct {
		ProgPath string
	}{
		ProgPath: agent.GetBinDir("gd-tools"),
	}
	content, err := templates.Parse(service, cfg.Verbose, data)
	if err != nil {
		return err
	}
	if err := os.WriteFile(service, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", service, err)
	}
	defer os.Remove(service)

	systemd := agent.GetEtcDir("systemd", "system", service)
	if _, err := cfg.LocalCommand(
		"rsync",
		cfg.RsyncFlags(),
		"--chown=root:root",
		"--chmod=0644",
		service,
		cfg.RootUser()+":"+systemd,
	); err != nil {
		return fmt.Errorf("failed to install %s: %w", service, err)
	}

	cmds := []string{
		"systemctl daemon-reexec",
		"systemctl daemon-reload",
		"systemctl enable gd-tools",
		"systemctl restart gd-tools",
	}
	if err := cfg.RemoteScript(cmds); err != nil {
		return fmt.Errorf("failed to enable %s: %w", service, err)
	}

	return nil
}
