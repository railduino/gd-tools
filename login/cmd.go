package login

import (
	"os"
	"os/exec"

	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Describe = `The login command opens a shell session on the server.`

var Command = &cli.Command{
	Name:        "login",
	Usage:       "Login to the production server as root",
	Description: Describe,
	Action:      Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	cmd := exec.Command("ssh", cfg.RootUser())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
