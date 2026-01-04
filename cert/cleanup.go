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

var CleanupCommand = &cli.Command{
	Name:  "cleanup",
	Usage: "Remove archived certificate directories (requires --force to delete)",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "force",
			Usage: "actually delete archive entries",
		},
		&cli.DurationFlag{
			Name:  "older-than",
			Usage: "only consider archives older than this duration (e.g. 720h = 30d)",
			Value: 90 * 24 * time.Hour,
		},
		&cli.IntFlag{
			Name:  "keep-last",
			Usage: "keep the most recent N archives per domain",
			Value: 1,
		},
	},
	Action: CleanupRun,
}

type archiveEntry struct {
	Domain string
	Path   string
	Mod    time.Time
}

func CleanupRun(c *cli.Context) error {
	if err := utils.EnsureHostDir(); err != nil {
		return err
	}

	archiveRoot := "acme-certs-archive"
	entries, err := os.ReadDir(archiveRoot)
	if err != nil {
		fmt.Println("No archived certificates found.")
		return nil
	}

	olderThan := c.Duration("older-than")
	keepLast := c.Int("keep-last")
	if keepLast < 0 {
		return fmt.Errorf("--keep-last must be >= 0")
	}

	var all []archiveEntry

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		domain := archiveNameToDomain(e.Name())
		if domain == "" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		all = append(all, archiveEntry{
			Domain: domain,
			Path:   filepath.Join(archiveRoot, e.Name()),
			Mod:    info.ModTime(),
		})
	}

	if len(all) == 0 {
		fmt.Println("No archived certificates found.")
		return nil
	}

	// group by domain
	byDomain := map[string][]archiveEntry{}
	for _, e := range all {
		byDomain[e.Domain] = append(byDomain[e.Domain], e)
	}

	now := time.Now().UTC()
	var toRemove []archiveEntry

	for domain, list := range byDomain {
		// newest first
		sort.Slice(list, func(i, j int) bool {
			return list[i].Mod.After(list[j].Mod)
		})

		for i, e := range list {
			// keep newest N
			if i < keepLast {
				continue
			}
			// respect age threshold
			if olderThan > 0 && now.Sub(e.Mod) < olderThan {
				continue
			}
			toRemove = append(toRemove, archiveEntry{
				Domain: domain,
				Path:   e.Path,
				Mod:    e.Mod,
			})
		}
	}

	if len(toRemove) == 0 {
		fmt.Println("Nothing to clean up.")
		return nil
	}

	// stable output
	sort.Slice(toRemove, func(i, j int) bool {
		if toRemove[i].Domain == toRemove[j].Domain {
			return toRemove[i].Mod.Before(toRemove[j].Mod)
		}
		return toRemove[i].Domain < toRemove[j].Domain
	})

	fmt.Printf("%-40s  %-19s  %s\n", "DOMAIN", "ARCHIVED (UTC)", "PATH")
	for _, e := range toRemove {
		fmt.Printf("%-40s  %-19s  %s\n",
			e.Domain,
			e.Mod.UTC().Format("2006-01-02 15:04:05"),
			e.Path,
		)
	}

	if !c.Bool("force") || c.Bool("dry") {
		fmt.Println("No files were removed. Use --force to delete the entries listed above.")
		return nil
	}

	for _, e := range toRemove {
		fmt.Printf("Deleting %s\n", e.Path)
		if err := os.RemoveAll(e.Path); err != nil {
			return err
		}
	}

	return nil
}
