package releases

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/railduino/gd-tools/templates"
)

const (
	ReleasesName = "releases.json"
)

// Load reads the release catalog from assets/releases.json.
func Load(verbose bool) (*Catalog, error) {
	releasesPath := filepath.Join("assets", ReleasesName)
	data, err := templates.Load(releasesPath, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", ReleasesName, err)
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", ReleasesName, err)
	}

	if err := catalog.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate %s: %w", ReleasesName, err)
	}

	return &catalog, nil
}
