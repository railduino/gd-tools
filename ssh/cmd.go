package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Describe = `The ssh command connects to the production server.

If no argument is given, it opens an interactive SSH session as root.
If one or more arguments are provided, it runs the given command remotely.`

var Command = &cli.Command{
	Name:        "ssh",
	Usage:       "Execute commands on the production server",
	Description: Describe,
	ArgsUsage:   "[<cmd> ...]",
	Action:      Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}
	remote := cfg.RootUser()

	// No arguments → interactive SSH session
	if c.NArg() == 0 {
		cmd := exec.Command("ssh", remote)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Arguments → execute remote shell command
	remoteCmd := strings.Join(c.Args().Slice(), " ")
	command := fmt.Sprintf("ssh %s '%s'", remote, remoteCmd)

	output, err := agent.RunShell([]string{command})
	fmt.Print(string(output))

	return err
}
