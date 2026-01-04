package cert

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var RestoreCommand = &cli.Command{
	Name:         "restore",
	Usage:        "Restore the latest archived certificate directory for a domain",
	ArgsUsage:    "<domain>",
	BashComplete: restoreBashComplete,
	Action:       RestoreRun,
}

func RestoreRun(c *cli.Context) error {
	if err := utils.EnsureHostDir(); err != nil {
		return err
	}

	domain := c.Args().First()
	if domain == "" {
		return fmt.Errorf("missing <domain>")
	}

	archiveRoot := "acme-certs-archive"
	entries, err := os.ReadDir(archiveRoot)
	if err != nil {
		return fmt.Errorf("no acme-certs-archive directory found")
	}

	type cand struct {
		path string
		mod  time.Time
	}
	var candidates []cand

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if archiveNameToDomain(name) != domain {
			continue
		}
		p := filepath.Join(archiveRoot, name)
		info, err := e.Info()
		if err != nil {
			continue
		}
		candidates = append(candidates, cand{path: p, mod: info.ModTime()})
	}

	if len(candidates) == 0 {
		return fmt.Errorf("no archived certificate found for domain %q", domain)
	}

	// newest first
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].mod.After(candidates[j].mod)
	})

	src := candidates[0].path
	dst := filepath.Join("acme-certs", domain)

	// safety: do not overwrite active certs
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("active certificate directory already exists: %s", dst)
	}

	if c.Bool("dry") {
		fmt.Printf("[dry] restore %s -> %s\n", src, dst)
		return nil
	}

	if err := os.MkdirAll("acme-certs", 0o755); err != nil {
		return err
	}

	if err := os.Rename(src, dst); err != nil {
		return err
	}

	fmt.Printf("Restored %s from %s\n", domain, filepath.Base(src))
	return nil
}

// restoreBashComplete completes domains that have archived certs.
func restoreBashComplete(c *cli.Context) {
	// Only complete in host dir
	if err := utils.EnsureHostDir(); err != nil {
		return
	}

	entries, err := os.ReadDir("acme-certs-archive")
	if err != nil {
		return
	}

	seen := map[string]struct{}{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		domain := archiveNameToDomain(e.Name())
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		fmt.Println(domain)
	}
}
