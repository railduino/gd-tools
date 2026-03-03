package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

const (
	BrevoName    = "brevo.json"
	BrevoBaseURL = "https://api.brevo.com/v3/senders/domains/"
)

type Brevo struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	API_Key  string `json:"api_key"`
	SMTP_ID  string `json:"smtp_id"`
	SMTP_Key string `json:"smtp_key"`
}

type BrevoRec struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	HostName string `json:"host_name"`
	Status   bool   `json:"status"`
}

type BrevoStatus struct {
	Domain        string `json:"domain"`
	Verified      bool   `json:"verified"`
	Authenticated bool   `json:"authenticated"`
	DNSRecords    struct {
		DKIMRecord  *BrevoRec `json:"dkim_record"`
		DKIM1Record *BrevoRec `json:"dkim1Record"`
		DKIM2Record *BrevoRec `json:"dkim2Record"`
		BrevoCode   *BrevoRec `json:"brevo_code"`
		DMARCRecord *BrevoRec `json:"dmarc_record"`
	} `json:"dns_records"`
}

func GetBrevo() (*Brevo, error) {
	content, err := os.ReadFile(BrevoName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", BrevoName, err)
	}

	var brv Brevo
	if err := json.Unmarshal(content, &brv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", BrevoName, err)
	}

	return &brv, nil
}

func ReadBrevo(c *cli.Context) (*Brevo, error) {
	content, err := os.ReadFile(BrevoName)
	if err != nil {
		if os.IsNotExist(err) {
			brv := Brevo{
				Server:   c.String("server"),
				Port:     c.Int("port"),
				API_Key:  c.String("api"),
				SMTP_ID:  c.String("id"),
				SMTP_Key: c.String("key"),
			}
			return &brv, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", BrevoName, err)
	}

	var brv Brevo
	if err := json.Unmarshal(content, &brv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", BrevoName, err)
	}

	return &brv, nil
}

func (brv *Brevo) Save() error {
	content, err := json.MarshalIndent(brv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", BrevoName, err)
	}

	existing, err := os.ReadFile(BrevoName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(BrevoName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", BrevoName, err)
	}

	return nil
}

func (dom *Domain) GetBrevo(apiKey string) (bool, error) {
	path := BrevoBaseURL + url.PathEscape(dom.Name)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("brevo: status=%d body=%s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var status BrevoStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return false, fmt.Errorf("brevo: parse failed: %w body=%s", err, strings.TrimSpace(string(body)))
	}
	fmt.Printf("Status: '%+v'\n", status)

	dom.BrevoValid = status.Verified && status.Authenticated

	if status.DNSRecords.BrevoCode != nil {
		dom.BrevoCode = status.DNSRecords.BrevoCode.Value
		fmt.Printf("Code: '%s'\n", dom.BrevoCode)
	}

	// Helper: extract DKIM selector from "<selector>._domainkey"
	addDKIM := func(rec *BrevoRec) {
		if rec == nil || !rec.Status {
			return
		}
		host := strings.TrimSpace(rec.HostName)
		val := strings.TrimSpace(rec.Value)
		typ := strings.ToUpper(strings.TrimSpace(rec.Type))

		const suffix = "._domainkey"
		if !strings.HasSuffix(host, suffix) {
			return
		}
		selector := strings.TrimSuffix(host, suffix)
		if selector == "" {
			return
		}
		dkim := DKIMRecord{Selector: selector}
		switch typ {
		case "CNAME":
			dkim.CNAME = val
		case "TXT":
			dkim.PubValue = val
		default:
			return
		}
		dom.AddDKIM(dkim)
	}

	// Brevo DKIM records
	addDKIM(raw.DNSRecords.DKIMRecord)
	addDKIM(raw.DNSRecords.DKIM1Record)
	addDKIM(raw.DNSRecords.DKIM2Record)

	return true, nil
}
