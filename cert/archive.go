package cert

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var ArchiveCommand = &cli.Command{
	Name:      "archive",
	Usage:     "Archive certificate directories (move to acme-certs-archive)",
	ArgsUsage: "[domain]",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.BoolFlag{
			Name:  "all",
			Usage: "archive all found certificate directories (requires --force)",
		},
		&cli.BoolFlag{
			Name:  "force",
			Usage: "required when using --all",
		},
	},
	BashComplete: archiveBashComplete,
	Action:       ArchiveRun,
}

func ArchiveRun(c *cli.Context) error {
	if err := utils.EnsureHostDir(); err != nil {
		return err
	}

	domain := c.Args().First()
	all := c.Bool("all")
	force := c.Bool("force")

	if all && !force {
		return fmt.Errorf("refusing --all without --force")
	}
	if !all && domain == "" {
		return fmt.Errorf("missing [domain] (or use --all --force)")
	}

	rows, err := findCertificates(".")
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Println("No ACME certificates found.")
		return nil
	}

	// Determine target domain dirs (dedupe)
	targets := make([]string, 0, len(rows))
	seen := map[string]struct{}{}

	for _, r := range rows {
		if all || r.Domain == domain {
			dir := certDomainDir(r.Path) // .../acme-certs/<domain>
			if _, ok := seen[dir]; ok {
				continue
			}
			seen[dir] = struct{}{}
			targets = append(targets, dir)
		}
	}

	if !all && len(targets) == 0 {
		return fmt.Errorf("no certificate directory found for domain %q", domain)
	}

	for _, dir := range targets {
		if c.Bool("dry") {
			fmt.Printf("[dry] archive %s\n", dir)
			continue
		}
		dst, err := archiveDir(dir)
		if err != nil {
			return err
		}
		fmt.Printf("Archived %s -> %s\n", dir, dst)
	}

	return nil
}

func archiveBashComplete(c *cli.Context) {
	if err := utils.EnsureHostDir(); err != nil {
		return
	}

	entries, err := os.ReadDir("acme-certs")
	if err != nil {
		return
	}

	for _, e := range entries {
		if e.IsDir() {
			fmt.Println(e.Name())
		}
	}
}

// archiveDir moves dir to a sibling archive location with timestamp suffix.
// Example:
//
//	<host>/acme-certs/<domain> -> <host>/acme-certs-archive/<domain>-YYYYmmdd-HHMMSS
func archiveDir(dir string) (string, error) {
	// dir is .../acme-certs/<domain>
	acmeDir := filepath.Dir(dir)     // .../acme-certs
	hostDir := filepath.Dir(acmeDir) // .../<host>
	archiveRoot := filepath.Join(hostDir, "acme-certs-archive")

	if err := os.MkdirAll(archiveRoot, 0o755); err != nil {
		return "", err
	}

	base := filepath.Base(dir) // <domain>
	ts := time.Now().UTC().Format("20060102-150405")
	dst := filepath.Join(archiveRoot, fmt.Sprintf("%s-%s", base, ts))

	return dst, os.Rename(dir, dst)
}
