package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

type certRow struct {
	Domain   string
	Path     string
	NotAfter time.Time
	DaysLeft int
}

var Describe = `The cert command manages ACME certificates.

It scans project folders for "acme-certs/<domain>/" and operates on the found certificates.

For detailed documentation and usage examples, see:
https://github.com/railduino/gd-tools/wiki/XX-Cert`

var Command = &cli.Command{
	Name:        "cert",
	Usage:       "List, archive, restore or sync ACME certificates",
	Description: Describe,
	Subcommands: []*cli.Command{
		ArchiveCommand,
		CleanupCommand,
		ListCommand,
		RestoreCommand,
		// TODO (later) SyncCommand,
	},
	Action: ListRun,
}

// certDomainDir returns the directory that contains the certificate
func certDomainDir(fullchainPath string) string {
	return filepath.Dir(fullchainPath)
}

// findCertificates collects all .../acme-certs/<domain>/fullchain.pem
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

// parsePEMNotAfter extracts NotAfter from the first CERTIFICATE in a PEM file
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

	return time.Time{}, fmt.Errorf("no certificate found in PEM")
}

// archiveNameToDomain reconstructs the domain from an archive
func archiveNameToDomain(name string) string {
	// Split on the last '-' before timestamp.
	// We chose <domain>-YYYYmmdd-HHMMSS.
	// Domains contain dots and hyphens, so we detect the trailing timestamp pattern loosely:
	// last two segments look like YYYYmmdd and HHMMSS.
	parts := strings.Split(name, "-")
	if len(parts) < 3 {
		return ""
	}
	// Expect last part HHMMSS, second last YYYYmmdd
	datePart := parts[len(parts)-2]
	timePart := parts[len(parts)-1]
	if len(datePart) != 8 || len(timePart) != 6 {
		return ""
	}
	// Re-join the rest as domain
	return strings.Join(parts[:len(parts)-2], "-")
}
