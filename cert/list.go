package cert

import (
	"fmt"
	"sort"

	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var ListCommand = &cli.Command{
	Name:   "list",
	Usage:  "Scan working directory for ACME certificates (fullchain.pem) and show expiry",
	Action: ListRun,
}

func ListRun(c *cli.Context) error {
	if err := utils.EnsureBaseOrHostDir(); err != nil {
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
