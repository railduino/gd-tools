package releases

import (
	"fmt"
)

// Validate checks the complete release catalog.
func (c *Catalog) Validate() error {
	for _, pr := range c.Products {
		if pr.Name == "" {
			return fmt.Errorf("product has no name")
		}

		if pr.Default == "" {
			return fmt.Errorf("product %q has no default release", pr.Name)
		}

		foundDefault := false

		for _, rel := range pr.Versions {
			if rel.Number == "" {
				return fmt.Errorf("release %s/%s has no version", pr.Name, rel.Number)
			}

			if rel.Download.DownloadURL == "" {
				return fmt.Errorf("release %s/%s has no download_url", pr.Name, rel.Number)
			}

			if rel.Download.Filename == "" {
				return fmt.Errorf("release %s/%s has no filename", pr.Name, rel.Number)
			}

			if rel.Download.MD5 == "" && rel.Download.SHA256 == "" && rel.Download.SHA512 == "" {
				return fmt.Errorf("release %s/%s has no checksum", pr.Name, rel.Number)
			}

			dir := rel.Download.Directory
			if dir == "" {
				dir = pr.Directory
			}

			bin := rel.Download.Binary
			if bin == "" {
				bin = pr.Binary
			}

			if dir != "" && bin != "" {
				return fmt.Errorf("release %s/%s defines both directory and binary", pr.Name, rel.Number)
			}

			if dir == "" && bin == "" {
				return fmt.Errorf("release %s/%s defines neither directory nor binary", pr.Name, rel.Number)
			}

			if rel.Number == pr.Default {
				foundDefault = true
			}
		}

		if !foundDefault {
			return fmt.Errorf("product %q default release %q not found", pr.Name, pr.Default)
		}
	}

	return nil
}
