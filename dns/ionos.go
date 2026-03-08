package dns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	IonosBaseURL = "https://api.hosting.ionos.com/dns/v1"
)

type IonosRecord struct {
	Name     string `json:"name"`
	ID       string `json:"id,omitempty"`
	Type     string `json:"type"`
	Value    string `json:"content"`
	TTL      int    `json:"ttl"`
	Prio     int    `json:"prio"`
	Disabled bool   `json:"disabled"`
	Text     string `json:"-"`
}

type IonosUpdate struct {
	Value    string `json:"content"`
	TTL      int    `json:"ttl"`
	Prio     int    `json:"prio"`
	Disabled bool   `json:"disabled"`
}

type IonosZone struct {
	Name    string        `json:"name"`
	ID      string        `json:"id"`
	Type    string        `json:"type"`
	Records []IonosRecord `json:"records"`
	rrSet   RRSet         `json:"-"`
}

type IonosProvider struct {
	AuthAPIToken string
	Client       *http.Client
}

func NewIonosProvider(apiToken string) (*IonosProvider, error) {
	if strings.TrimSpace(apiToken) == "" {
		return nil, fmt.Errorf("ionos: api token required")
	}

	return &IonosProvider{
		AuthAPIToken: apiToken,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (rec *IonosRecord) String() string {
	content, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return "failed to Marshal IonosRecord"
	}

	return string(content)
}

func (zone *IonosZone) String() string {
	content, err := json.MarshalIndent(zone, "", "  ")
	if err != nil {
		return "failed to Marshal IonosZone"
	}

	return string(content)
}

func ionosFQDN(zone, name string) string {
	zone = normalizeZone(zone)
	name = normalizeName(name)

	if name == "@" {
		return zone
	}
	return name + "." + zone
}

func (p *IonosProvider) UpsertRRSet(ctx context.Context, zone string, rrIn RRSet) (string, error) {
	zone = normalizeZone(zone)
	rr, err := normalizeRRSet(rrIn)
	if err != nil {
		return "", err
	}

	theZone, err := p.getZoneData(ctx, zone, rr)
	if err != nil {
		return "", err
	}

	if status := p.checkSame(theZone); status != "" {
		return status, nil
	}
	var result []string

	// Delete existing records (except TXT for Apex and CAA)
	if (rr.Type != RR_TXT && rr.Type != RR_CAA) || rr.Name != "@" {
		for i := range theZone.Records {
			url := IonosBaseURL + "/zones/" + theZone.ID + "/records/" + theZone.Records[i].ID
			if _, err := p.sendToAPI(ctx, "DELETE", url, nil); err != nil {
				return "", err
			}
			result = append(result,
				fmt.Sprintf("✅ DNS (%s) deleted %v for %s: %s", theZone.Name, rr.Type, rr.Name, theZone.Records[i].Value),
			)
		}
	}

	// Create desired records
	for i := range rr.Records {
		rec := IonosRecord{
			Name:     ionosFQDN(theZone.Name, rr.Name),
			Type:     strings.ToUpper(string(rr.Type)),
			Value:    rr.Records[i].Value,
			TTL:      rr.TTL,
			Prio:     rr.Records[i].Prio,
			Disabled: false,
		}

		body, err := json.Marshal([]IonosRecord{rec})
		if err != nil {
			return "", err
		}

		url := IonosBaseURL + "/zones/" + theZone.ID + "/records"
		if _, err := p.sendToAPI(ctx, "POST", url, body); err != nil {
			return "", err
		}

		result = append(result,
			fmt.Sprintf("✅ DNS (%s) created %v for %s: %s", theZone.Name, rr.Type, rr.Name, rr.Records[i].Value),
		)
	}

	return strings.Join(result, "\n"), nil
}

// checkSame compares the desired RRSet with the currently existing records
// for the same (zone, name, type) and determines whether an update is a no-op.
//
// Contract / Assumptions:
//   - theZone.Records contains only records matching rrSet.Name and rrSet.Type.
//   - both theZone.Records and rrSet.Records are deterministically sorted
//     using the same ordering rules (currently: prio asc, value asc).
//   - rrSet.Records are already normalized for provider-specific semantics
//     (e.g. trailing dots, TXT wire format, MX text representation).
//   - IonosRecord.Text and RRRecord.Text represent the canonical comparison
//     form for the record content.
//   - rrSet.TTL is the authoritative TTL and must match the per-record TTL
//     returned by the provider.
//
// Under these assumptions, a positional comparison is sufficient and stable.
func (p *IonosProvider) checkSame(theZone *IonosZone) string {
	rr := theZone.rrSet

	if len(theZone.Records) != len(rr.Records) {
		return ""
	}

	for i := range rr.Records {
		if theZone.Records[i].TTL != rr.TTL {
			return ""
		}
		if theZone.Records[i].Text != rr.Records[i].Text {
			return ""
		}
	}

	count := "1 record"
	if n := len(rr.Records); n != 1 {
		count = fmt.Sprintf("%d records", n)
	}
	return fmt.Sprintf("✅ DNS (%s) no-op %v for %s: %s", theZone.Name, rr.Type, rr.Name, count)
}

func (p *IonosProvider) getZoneData(ctx context.Context, zone string, rr RRSet) (*IonosZone, error) {
	zone = normalizeZone(zone)

	zoneID, err := p.getZoneID(ctx, zone)
	if err != nil {
		return nil, err
	}

	var theZone IonosZone
	url := IonosBaseURL + "/zones/" + zoneID
	data, err := p.sendToAPI(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &theZone); err != nil {
		return nil, err
	}

	wantType := strings.ToUpper(string(rr.Type))
	wantName := ionosFQDN(zone, rr.Name)
	var records []IonosRecord
	for _, zoneRec := range theZone.Records {
		if zoneRec.Disabled {
			continue
		}
		if strings.ToUpper(zoneRec.Type) != wantType {
			continue
		}
		if normalizeZone(zoneRec.Name) != wantName {
			continue
		}
		if zoneRec.Value = strings.TrimSpace(zoneRec.Value); zoneRec.Value == "" {
			continue
		}
		switch rr.Type {
		case RR_CNAME:
			host := strings.TrimSuffix(zoneRec.Value, ".")
			zoneRec.Value = host
			zoneRec.Text = host + "."
		case RR_MX:
			host := strings.TrimSuffix(zoneRec.Value, ".")
			zoneRec.Value = host
			zoneRec.Text = fmt.Sprintf("%d %s.", zoneRec.Prio, host)
		case RR_TXT:
			zoneRec.Text = canonTXTWire(zoneRec.Value)
		default:
			zoneRec.Text = zoneRec.Value
		}
		records = append(records, zoneRec)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Prio != records[j].Prio {
			return records[i].Prio < records[j].Prio
		}
		return records[i].Value < records[j].Value
	})

	theZone.Records = records // filtered by typ and name, already sorted
	theZone.rrSet = rr

	return &theZone, nil
}

func (p *IonosProvider) getZoneID(ctx context.Context, name string) (string, error) {
	url := IonosBaseURL + "/zones"
	data, err := p.sendToAPI(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	var zones []IonosZone
	if err := json.Unmarshal(data, &zones); err != nil {
		return "", err
	}

	for _, zone := range zones {
		if name == zone.Name {
			return zone.ID, nil
		}
	}

	return "", fmt.Errorf("ionos: ID not found for zone %s", name)
}

func (p *IonosProvider) sendToAPI(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.AuthAPIToken)

	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Keep body because IONOS returns useful JSON/text errors.
		return nil, fmt.Errorf("ionos: HTTP %d %s: %s", resp.StatusCode, resp.Status, strings.TrimSpace(string(data)))
	}

	return data, nil
}
