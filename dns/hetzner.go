package dns

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type HetznerProvider struct {
	client *hcloud.Client
	zones  map[string]*hcloud.Zone
	mu     sync.Mutex
}

func toHcloudType(t RRSetType) (hcloud.ZoneRRSetType, error) {
	switch t {
	case RR_A:
		return hcloud.ZoneRRSetTypeA, nil
	case RR_AAAA:
		return hcloud.ZoneRRSetTypeAAAA, nil
	case RR_CNAME:
		return hcloud.ZoneRRSetTypeCNAME, nil
	case RR_MX:
		return hcloud.ZoneRRSetTypeMX, nil
	case RR_TXT:
		return hcloud.ZoneRRSetTypeTXT, nil
	case RR_CAA:
		return hcloud.ZoneRRSetTypeCAA, nil
	default:
		return "", fmt.Errorf("unsupported RRSetType %q", t)
	}
}

func NewHetznerProvider(apiToken string) (*HetznerProvider, error) {
	if strings.TrimSpace(apiToken) == "" {
		return nil, fmt.Errorf("hetzner: api token required")
	}

	client := hcloud.NewClient(
		hcloud.WithToken(apiToken),
	)

	return &HetznerProvider{
		client: client,
		zones:  make(map[string]*hcloud.Zone),
	}, nil
}

func (p *HetznerProvider) getZone(ctx context.Context, name string) (*hcloud.Zone, error) {
	name = strings.ToLower(strings.TrimSuffix(name, "."))

	p.mu.Lock()
	if z, ok := p.zones[name]; ok {
		p.mu.Unlock()
		return z, nil
	}
	p.mu.Unlock()

	z, _, err := p.client.Zone.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("hetzner: get zone %q failed: %w", name, err)
	}
	if z == nil {
		return nil, fmt.Errorf("hetzner: zone %q not visible for token/project", name)
	}

	p.mu.Lock()
	p.zones[name] = z
	p.mu.Unlock()

	return z, nil
}

func (p *HetznerProvider) UpsertRRSet(ctx context.Context, zone string, rrIn RRSet) (string, error) {
	zone = normalizeZone(zone)
	rr, err := normalizeRRSet(rrIn)
	if err != nil {
		return "", err
	}

	z, err := p.getZone(ctx, zone)
	if err != nil {
		return "", err
	}

	ht, err := toHcloudType(rr.Type)
	if err != nil {
		return "", err
	}

	// convert []RRecord -> []hcloud.ZoneRRSetRecord
	wanted := make([]hcloud.ZoneRRSetRecord, 0, len(rr.Records))
	for _, v := range rr.Records {
		wanted = append(wanted, hcloud.ZoneRRSetRecord{Value: v.Text})
	}

	// Get existing
	currRRSet, _, err := p.client.Zone.GetRRSetByNameAndType(ctx, z, rr.Name, ht)
	if err != nil && !hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		return "", fmt.Errorf("hetzner: get rrset %s %s %s failed: %w", zone, rr.Name, rr.Type, err)
	}
	var existing []hcloud.ZoneRRSetRecord
	if currRRSet != nil {
		for _, rec := range currRRSet.Records {
			value := rec.Value
			if rr.Type == "TXT" {
				value = canonTXTWire(value)
			}
			existing = append(existing, hcloud.ZoneRRSetRecord{Value: value})
		}
		sort.Slice(existing, func(i, j int) bool {
			return existing[i].Value < existing[j].Value
		})
	}

	// Create if missing
	if len(existing) == 0 || hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		_, _, err := p.client.Zone.CreateRRSet(ctx, z, hcloud.ZoneRRSetCreateOpts{
			Name:    rr.Name,
			Type:    ht,
			TTL:     &rr.TTL,
			Records: wanted,
		})
		if err != nil {
			return "", fmt.Errorf("hetzner: create rrset %s %s %s failed: %w", zone, rr.Name, rr.Type, err)
		}
		return fmt.Sprintf("✅ DNS (%s) created %s for %s: %v", zone, rr.Type, rr.Name, rr.Records), nil
	}

	// Check if update is necessary
	if hetznerIsEqual(existing, wanted) {
		return fmt.Sprintf("✅ DNS (%s) no-op %s for %s: %v", zone, rr.Type, rr.Name, rr.Records), nil
	}

	// Update if different
	_, _, err = p.client.Zone.SetRRSetRecords(ctx, currRRSet, hcloud.ZoneRRSetSetRecordsOpts{
		Records: wanted,
	})
	if err != nil {
		return "", fmt.Errorf("hetzner: update rrset %s %s %s failed: %w", zone, rr.Name, rr.Type, err)
	}

	return fmt.Sprintf("✅ DNS (%s) updated %s for %s: %v", zone, rr.Type, rr.Name, rr.Records), nil
}

func hetznerIsEqual(a, b []hcloud.ZoneRRSetRecord) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}
