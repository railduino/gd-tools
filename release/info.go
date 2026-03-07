package release

import (
	"fmt"
	"sort"

	"github.com/railduino/gd-tools/releases"
	"github.com/urfave/cli/v2"
)

var InfoDescribe = `The info command shows the embedded release catalog.

Without arguments, it shows all products.
With one argument, it shows only the selected product.

Examples:
  gdt release info
  gdt release info nextcloud
  gdt release info --products`

var InfoCommand = &cli.Command{
	Name:        "info",
	Usage:       "Show embedded release catalog",
	Description: InfoDescribe,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "products",
			Usage: "show only product names",
		},
	},
	Action: InfoRun,
}

func InfoRun(c *cli.Context) error {
	catalog, err := releases.Load(false)
	if err != nil {
		return err
	}

	if c.Bool("products") {
		var names []string
		for _, pr := range catalog.Products {
			names = append(names, pr.Name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Println(name)
		}
		return nil
	}

	name := ""
	if c.NArg() == 1 {
		name = c.Args().First()
	}

	space := false
	for _, pr := range catalog.Products {
		if name != "" && name != pr.Name {
			continue
		}
		if space {
			fmt.Println()
		}
		space = true

		for _, line := range pr.Info() {
			fmt.Println(line)
		}
	}

	return nil
}
