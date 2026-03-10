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
	BrevoServer  = "smtp-relay.brevo.com"
	BrevoPort    = 587
	BrevoBaseURL = "https://api.brevo.com/v3/senders/domains/"
)

type Brevo struct {
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

type BrevoData struct {
	Domain        string `json:"domain"`
	Verified      bool   `json:"verified"`
	Authenticated bool   `json:"authenticated"`
	DNSRecords    struct {
		DKIMRecord  *BrevoRec `json:"dkim_record"`
		DKIM1Record *BrevoRec `json:"dkim1Record"`
		DKIM2Record *BrevoRec `json:"dkim2Record"`
		BrevoCode   *BrevoRec `json:"brevo_code"`
		DMARCRecord *BrevoRec `json:"dmarc_record"` // not used
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

func BrevoSASL() (string, error) {
	brv, err := GetBrevo()
	if err != nil || brv == nil {
		return "", err
	}

	if brv.SMTP_ID == "" || brv.SMTP_Key == "" {
		return "", fmt.Errorf("Brevo is missing SMTP credentials")
	}

	return fmt.Sprintf("[%s]:%d %s:%s",
		BrevoServer,
		BrevoPort,
		brv.SMTP_ID,
		brv.SMTP_Key,
	), nil
}

func BrevoTarget() (string, int) {
	return BrevoServer, BrevoPort
}

func ReadBrevo(c *cli.Context) (*Brevo, error) {
	content, err := os.ReadFile(BrevoName)
	if err != nil {
		if os.IsNotExist(err) {
			brv := Brevo{
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

func (dom *Domain) BrevoUpdate(apiKey string) (bool, error) {
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

	var data BrevoData
	if err := json.Unmarshal(body, &data); err != nil {
		return false, fmt.Errorf("brevo: parse failed: %w body=%s", err, strings.TrimSpace(string(body)))
	}

	if code := data.DNSRecords.BrevoCode; code != nil && code.Value != "" {
		dom.BrevoCode = code.Value
	}

	// Helper: extract DKIM selector from "<selector>._domainkey"
	addDKIM := func(rec *BrevoRec) {
		if rec == nil {
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
		dkim := DKIM{Selector: selector}
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
	addDKIM(data.DNSRecords.DKIMRecord)
	addDKIM(data.DNSRecords.DKIM1Record)
	addDKIM(data.DNSRecords.DKIM2Record)

	// Looks good, add Brevo to the SPF list
	dom.AddSPF("include:spf.brevo.com")

	return true, nil
}

func (dom *Domain) GetBrevoStatus(apiKey string) (string, error) {
	path := BrevoBaseURL + url.PathEscape(dom.Name)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return "missing", nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("brevo: status=%d body=%s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var data BrevoData
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("brevo: parse failed: %w body=%s", err, strings.TrimSpace(string(body)))
	}

	if data.Authenticated {
		return "authenticated", nil
	}

	if data.Verified {
		return "verified", nil
	}

	return "pending", nil
}
