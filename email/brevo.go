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

type BrevoData struct {
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

	if dmarc := data.DNSRecords.DMARCRecord; dmarc != nil && dmarc.Value != "" {
		dom.DMARC = dmarc.Value
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

func (dom *Domain) BrevoStatus(apiKey string) (string, bool, error) {
	path := BrevoBaseURL + url.PathEscape(dom.Name)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return "   --- Brevo:missing", false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("brevo: status=%d body=%s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var data BrevoData
	if err := json.Unmarshal(body, &data); err != nil {
		return "", false, fmt.Errorf("brevo: parse failed: %w body=%s", err, strings.TrimSpace(string(body)))
	}

	if data.Authenticated {
		return "   +++ Brevo:Authenticated", true, nil
	}

	if data.Verified {
		return "   *** Brevo:Verified", false, nil
	}

	return "   ??? Brevo:Pending", false, nil
}

func CheckBrevo() (*Brevo, error) {
	brevo, err := GetBrevo()
	if err != nil {
		return nil, err
	}
	if brevo == nil || brevo.API_Key == "" {
		return nil, nil
	}
	if brevo.SMTP_ID == "" || brevo.SMTP_Key == "" {
		return nil, nil
	}
	if brevo.Server == "" || brevo.Port == 0 {
		return nil, nil
	}

	domainList, _, err := GetDomains(nil)
	if err != nil {
		return nil, err
	}

	var brevoMissing []string
	for _, dom := range domainList.Domains {
		_, valid, err := dom.BrevoStatus(brevo.API_Key)
		if err != nil {
			return nil, err
		}
		if !valid {
			brevoMissing = append(brevoMissing, dom.Name)
		}
	}

	if len(brevoMissing) > 0 {
		return nil, fmt.Errorf("missing Brevo domains: %s", strings.Join(brevoMissing, ", "))
	}

	return brevo, nil
}
