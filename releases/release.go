package releases

import (
	"fmt"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/utils"
)

// Release describes one concrete release entry in the catalog.
type Release struct {
	Number   string         `json:"number"`
	Series   string         `json:"series,omitempty"`
	Download agent.Download `json:"download"`
}

// Product describes a given product (like Nextcloud, WordPress, etc.).
type Product struct {
	Name      string    `json:"name"`
	SourceURL string    `json:"source_url,omitempty"`
	Default   string    `json:"default"`
	Directory string    `json:"directory,omitempty"`
	Binary    string    `json:"binary,omitempty"`
	Versions  []Release `json:"versions"`
}

// Catalog is the root structure of releases.json.
type Catalog struct {
	Products []Product `json:"products"`
}

// Get returns a specific release for one product.
// Return the default release if num is empty.
func (c *Catalog) Get(name, num string) (*Product, *Release, error) {
	for i := range c.Products {
		pr := &c.Products[i]
		if pr.Name != name {
			continue
		}

		rel, err := pr.Get(num)
		if err != nil {
			return nil, nil, err
		}

		return pr, rel, nil
	}

	return nil, nil, fmt.Errorf("product %q not found", name)
}

// Get returns a specific product release.
// Return the default release if num is empty.
func (pr *Product) Get(num string) (*Release, error) {
	if num == "" {
		num = pr.Default
	}

	for i := range pr.Versions {
		rel := &pr.Versions[i]
		if rel.Number == num {
			if rel.Download.Directory == "" {
				rel.Download.Directory = pr.Directory
			}
			if rel.Download.Binary == "" {
				rel.Download.Binary = pr.Binary
			}
			return rel, nil
		}
	}

	return nil, fmt.Errorf("release %q not found for product %q", num, pr.Name)
}

// Return info for the given product
func (pr *Product) Info() string {
	var lb utils.LineBuffer

	lb.Addf("Product:    %s", pr.Name)
	if pr.SourceURL != "" {
		lb.Addf("Source URL: %s", pr.SourceURL)
	}

	lb.Add("Known releases:")
	for _, rel := range pr.Versions {
		line := rel.Number
		if rel.Number == pr.Default {
			line += " (default)"
		}
		lb.Addf("  - %s", line)

		lb.Addf("    Version:    %s", rel.Number)

		if rel.Series != "" {
			lb.Addf("    Series:     %s", rel.Series)
		}

		lb.Addf("    File:       %s", rel.Download.Filename)

		dir := rel.Download.Directory
		if dir == "" {
			dir = pr.Directory
		}
		if dir != "" {
			lb.Addf("    Directory:  %s", dir)
		}

		bin := rel.Download.Binary
		if bin == "" {
			bin = pr.Binary
		}
		if bin != "" {
			lb.Addf("    Binary:     %s", bin)
		}

		if rel.Download.MD5 != "" {
			lb.Addf("    MD5:        %s", rel.Download.MD5)
		}
		if rel.Download.SHA256 != "" {
			lb.Addf("    SHA256:     %s", rel.Download.SHA256)
		}
		if rel.Download.SHA512 != "" {
			lb.Addf("    SHA512:     %s", rel.Download.SHA512)
		}

		lb.Addf("    URL:        %s", rel.Download.DownloadURL)
	}

	return lb.Text()
}
