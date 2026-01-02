package cert

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

// ListCommand registers the "list" subcommand.
var ListCommand = &cli.Command{
	Name:   "list",
	Usage:  "Scan working directory for ACME certificates (fullchain.pem) and show expiry",
	Action: ListRun,
}

type certRow struct {
	Domain   string
	Path     string
	NotAfter time.Time
	DaysLeft int
}

// ListRun executes the "list" subcommand.
func ListRun(c *cli.Context) error {
	// Guard: ensure we are in a valid gd-tools base dir
	if err := utils.EnsureBaseDir(); err != nil {
		return err
	}

	rows, err := findCertificates(".")
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Println("No ACME certificates found.")
		return nil
	}

	// Sort by expiry (soonest last), then by domain for stability
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].NotAfter.Equal(rows[j].NotAfter) {
			return rows[i].Domain < rows[j].Domain
		}
		return rows[i].NotAfter.After(rows[j].NotAfter)
	})

	// Print table
	fmt.Printf("%-40s  %-19s  %-4s  %s\n", "DOMAIN", "EXPIRES (UTC)", "DAYS", "PATH")
	for _, r := range rows {
		fmt.Printf("%-40s  %-19s  %4d  %s\n",
			r.Domain, r.NotAfter.UTC().Format("2006-01-02 15:04:05"), r.DaysLeft, r.Path)
	}
	return nil
}

// findCertificates collects all .../acme-certs/<domain>/fullchain.pem under root.
func findCertificates(root string) ([]certRow, error) {
	var out []certRow

	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.Name() != "fullchain.pem" {
			return nil
		}

		// We expect: .../acme-certs/<domain>/fullchain.pem
		domainDir := filepath.Dir(p)
		acmeDir := filepath.Dir(domainDir)
		if filepath.Base(acmeDir) != "acme-certs" {
			return nil
		}
		domain := filepath.Base(domainDir)

		notAfter, err := parsePEMNotAfter(p)
		if err != nil {
			fmt.Printf("failed to parse %s: %v\n", p, err)
			return nil
		}
		out = append(out, certRow{
			Domain:   domain,
			Path:     p,
			NotAfter: notAfter,
			DaysLeft: int(time.Until(notAfter).Hours() / 24),
		})
		return nil
	})

	return out, err
}

// parsePEMNotAfter extracts NotAfter from the first CERTIFICATE in a PEM file.
func parsePEMNotAfter(pemPath string) (time.Time, error) {
	data, err := os.ReadFile(pemPath)
	if err != nil {
		return time.Time{}, err
	}
	var block *pem.Block
	rest := data
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				continue
			}
			return cert.NotAfter, nil
		}
	}
	return time.Time{}, errors.New("no certificate found in PEM")
}
